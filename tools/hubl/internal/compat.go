package internal

import (
	"context"
	"fmt"
	"io"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1beta1 "cosmossdk.io/api/cosmos/base/reflection/v1beta1"
)

// loadFileDescriptorsGRPCReflection attempts to load the file descriptor set using gRPC reflection when cosmos.reflection.v1
// is unavailable.
func loadFileDescriptorsGRPCReflection(ctx context.Context, client *grpc.ClientConn) (*descriptorpb.FileDescriptorSet, error) {
	fmt.Printf("This chain does not support cosmos.reflection.v1 yet... attempting to use a fallback. Some features may be unsupported and it may not be possible to read all data.\n")

	var interfaceImplNames []string
	cosmosReflectBetaClient := reflectionv1beta1.NewReflectionServiceClient(client)
	interfacesRes, err := cosmosReflectBetaClient.ListAllInterfaces(ctx, &reflectionv1beta1.ListAllInterfacesRequest{})
	if err == nil {
		for _, iface := range interfacesRes.InterfaceNames {
			implRes, err := cosmosReflectBetaClient.ListImplementations(ctx, &reflectionv1beta1.ListImplementationsRequest{
				InterfaceName: iface,
			})
			if err == nil {
				interfaceImplNames = append(interfaceImplNames, implMsgNameCleanup(implRes.ImplementationMessageNames)...)
			}
		}
	}

	reflectClient, err := grpc_reflection_v1alpha.NewServerReflectionClient(client).ServerReflectionInfo(ctx)
	if err != nil {
		return nil, err
	}

	fdMap := map[string]*descriptorpb.FileDescriptorProto{}
	waitListServiceRes := make(chan *grpc_reflection_v1alpha.ListServiceResponse) //nolint:staticcheck // we want to use the deprecated field
	waitc := make(chan struct{})
	go func() {
		for {
			in, err := reflectClient.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}
			if err != nil {
				panic(err)
			}

			switch res := in.MessageResponse.(type) {
			case *grpc_reflection_v1alpha.ServerReflectionResponse_ErrorResponse:
				panic(err)
			case *grpc_reflection_v1alpha.ServerReflectionResponse_ListServicesResponse:
				waitListServiceRes <- res.ListServicesResponse //nolint:staticcheck // we want to use the deprecated field
			case *grpc_reflection_v1alpha.ServerReflectionResponse_FileDescriptorResponse:
				_ = processFileDescriptorsResponse(res, fdMap)
			}
		}
	}()

	if err = reflectClient.Send(&grpc_reflection_v1alpha.ServerReflectionRequest{ //nolint:staticcheck // we want to use the deprecated field
		MessageRequest: &grpc_reflection_v1alpha.ServerReflectionRequest_ListServices{},
	}); err != nil {
		return nil, err
	}

	listServiceRes := <-waitListServiceRes

	for _, response := range listServiceRes.Service { //nolint:staticcheck // we want to use the deprecated field
		err = reflectClient.Send(&grpc_reflection_v1alpha.ServerReflectionRequest{ //nolint:staticcheck // we want to use the deprecated field
			MessageRequest: &grpc_reflection_v1alpha.ServerReflectionRequest_FileContainingSymbol{
				FileContainingSymbol: response.Name, //nolint:staticcheck // we want to use the deprecated field
			},
		})
		if err != nil {
			return nil, err
		}
	}

	for _, msgName := range interfaceImplNames {
		err = reflectClient.Send(&grpc_reflection_v1alpha.ServerReflectionRequest{ //nolint:staticcheck // we want to use the deprecated field
			MessageRequest: &grpc_reflection_v1alpha.ServerReflectionRequest_FileContainingSymbol{
				FileContainingSymbol: msgName,
			},
		})
		if err != nil {
			return nil, err
		}
	}

	if err = reflectClient.CloseSend(); err != nil {
		return nil, err
	}

	<-waitc

	// we loop through all the file descriptor dependencies to capture any file descriptors we haven't loaded yet
	cantFind := map[string]bool{}
	for {
		missing := missingFileDescriptors(fdMap, cantFind)
		if len(missing) == 0 {
			break
		}

		err = addMissingFileDescriptors(ctx, client, fdMap, missing)
		if err != nil {
			return nil, err
		}

		// mark all deps that we aren't able to resolve as can't find, so we don't keep looping and get a 429 error
		for _, dep := range missing {
			if fdMap[dep] == nil {
				cantFind[dep] = true
			}
		}
	}

	for dep := range cantFind {
		fmt.Printf("Warning: can't find %s.\n", dep)
	}

	fdSet := &descriptorpb.FileDescriptorSet{}
	for _, descriptorProto := range fdMap {
		fdSet.File = append(fdSet.File, descriptorProto)
	}

	return fdSet, nil
}

func processFileDescriptorsResponse(res *grpc_reflection_v1alpha.ServerReflectionResponse_FileDescriptorResponse, fdMap map[string]*descriptorpb.FileDescriptorProto) error {
	for _, bz := range res.FileDescriptorResponse.FileDescriptorProto { //nolint:staticcheck // we want to use the deprecated field
		fd := &descriptorpb.FileDescriptorProto{}
		err := proto.Unmarshal(bz, fd)
		if err != nil {
			return fmt.Errorf("error unmarshalling file descriptor: %w", err)
		}

		fdMap[fd.GetName()] = fd
	}

	return nil
}

func missingFileDescriptors(fdMap map[string]*descriptorpb.FileDescriptorProto, cantFind map[string]bool) []string {
	var missing []string
	for _, descriptorProto := range fdMap {
		for _, dep := range descriptorProto.Dependency {
			if fdMap[dep] == nil && !cantFind[dep] /* skip deps we've marked as can't find */ {
				missing = append(missing, dep)
			}
		}
	}
	return missing
}

func addMissingFileDescriptors(ctx context.Context, client *grpc.ClientConn, fdMap map[string]*descriptorpb.FileDescriptorProto, missingFiles []string) error {
	reflectClient, err := grpc_reflection_v1alpha.NewServerReflectionClient(client).ServerReflectionInfo(ctx)
	if err != nil {
		return err
	}

	waitc := make(chan struct{})
	go func() {
		for {
			in, err := reflectClient.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}
			if err != nil {
				panic(err)
			}

			if res, ok := in.MessageResponse.(*grpc_reflection_v1alpha.ServerReflectionResponse_FileDescriptorResponse); ok {
				_ = processFileDescriptorsResponse(res, fdMap)
			}
		}
	}()

	for _, file := range missingFiles {
		err = reflectClient.Send(&grpc_reflection_v1alpha.ServerReflectionRequest{ //nolint:staticcheck // we want to use the deprecated field
			MessageRequest: &grpc_reflection_v1alpha.ServerReflectionRequest_FileByFilename{
				FileByFilename: file,
			},
		})
		if err != nil {
			return err
		}
	}

	err = reflectClient.CloseSend()
	if err != nil {
		return err
	}

	<-waitc
	return nil
}

func guessAutocli(files *protoregistry.Files) *autocliv1.AppOptionsResponse {
	fmt.Printf("This chain does not support autocli directly yet. Using some default mappings in the meantime to support a subset of the available services.\n")
	res := map[string]*autocliv1.ModuleOptions{}
	files.RangeFiles(func(descriptor protoreflect.FileDescriptor) bool {
		services := descriptor.Services()
		n := services.Len()
		for i := 0; i < n; i++ {
			service := services.Get(i)
			serviceName := service.FullName()
			mapping, ok := defaultAutocliMappings[serviceName]
			if ok {
				parts := strings.Split(mapping, " ")
				numParts := len(parts)
				if numParts < 2 || numParts > 3 {
					fmt.Printf("Warning: bad mapping %q found for %q\n", mapping, serviceName)
					continue
				}

				modOpts := res[parts[0]]
				if modOpts == nil {
					modOpts = &autocliv1.ModuleOptions{}
					res[parts[0]] = modOpts
				}

				switch parts[1] {
				case "query":
					if modOpts.Query == nil {
						modOpts.Query = &autocliv1.ServiceCommandDescriptor{
							SubCommands: map[string]*autocliv1.ServiceCommandDescriptor{},
						}
					}
					if numParts == 3 {
						modOpts.Query.SubCommands[parts[2]] = &autocliv1.ServiceCommandDescriptor{Service: string(serviceName)}
					} else {
						modOpts.Query.Service = string(serviceName)
					}
				case "tx":
					if modOpts.Tx == nil {
						modOpts.Tx = &autocliv1.ServiceCommandDescriptor{
							SubCommands: map[string]*autocliv1.ServiceCommandDescriptor{},
						}
					}
					if numParts == 3 {
						modOpts.Tx.SubCommands[parts[2]] = &autocliv1.ServiceCommandDescriptor{Service: string(serviceName)}
					} else {
						modOpts.Tx.Service = string(serviceName)
					}
				default:
					fmt.Printf("Warning: bad mapping %q found for %q\n", mapping, serviceName)
					continue
				}
			}
		}
		return true
	})

	return &autocliv1.AppOptionsResponse{ModuleOptions: res}
}

// Removes the first character "/" from the received name
func implMsgNameCleanup(implMessages []string) (cleanImplMessages []string) {
	for _, implMessage := range implMessages {
		if len(implMessage) >= 1 && implMessage[0] == '/' {
			cleanImplMessages = append(cleanImplMessages, implMessage[1:])
		} else {
			cleanImplMessages = append(cleanImplMessages, implMessage)
		}
	}

	return cleanImplMessages
}

var defaultAutocliMappings = map[protoreflect.FullName]string{
	"cosmos.auth.v1beta1.Query":         "auth query",
	"cosmos.authz.v1beta1.Query":        "authz query",
	"cosmos.bank.v1beta1.Query":         "bank query",
	"cosmos.distribution.v1beta1.Query": "distribution query",
	"cosmos.evidence.v1.Query":          "evidence query",
	"cosmos.feegrant.v1beta1.Query":     "feegrant query",
	"cosmos.gov.v1.Query":               "gov query",
	"cosmos.gov.v1beta1.Query":          "gov query v1beta1",
	"cosmos.group.v1.Query":             "group query",
	"cosmos.mint.v1beta1.Query":         "mint query",
	"cosmos.params.v1beta1.Query":       "params query",
	"cosmos.slashing.v1beta1.Query":     "slashing query",
	"cosmos.staking.v1beta1.Query":      "staking query",
	"cosmos.upgrade.v1.Query":           "upgrade query",
}
