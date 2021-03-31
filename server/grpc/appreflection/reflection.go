package appreflection

import (
	"context"
	"fmt"

	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

type Config struct {
	SigningModes      []string
	ChainID           string
	SdkConfig         *sdk.Config
	InterfaceRegistry codectypes.InterfaceRegistry
}

// Register registers the cosmos sdk reflection service
// to the provided *grpc.Server given a Config
func Register(srv *grpc.Server, conf Config) error {

	reflectionServer, err := newReflectionServiceServer(srv, conf)
	if err != nil {
		return err
	}
	RegisterReflectionServiceServer(srv, reflectionServer)
	return nil
}

type reflectionServiceServer struct {
	desc                  *AppDescriptor
	interfacesList        []string
	interfaceImplementers map[string][]string
}

func newReflectionServiceServer(grpcSrv *grpc.Server, conf Config) (reflectionServiceServer, error) {
	// set chain descriptor
	chainDescriptor := &ChainDescriptor{Id: conf.ChainID}
	// set configuration descriptor
	configurationDescriptor := &ConfigurationDescriptor{
		Bech32AccountAddressPrefix: conf.SdkConfig.GetBech32AccountAddrPrefix(),
	}
	// set codec descriptor
	codecDescriptor, err := newCodecDescriptor(conf.InterfaceRegistry)
	if err != nil {
		return reflectionServiceServer{}, fmt.Errorf("unable to create codec descriptor: %w", err)
	}
	// set query service descriptor
	queryServiceDescriptor := newQueryServiceDescriptor(grpcSrv)
	// set deliver descriptor
	txDescriptor, err := newTxDescriptor(conf.InterfaceRegistry, conf.SigningModes)
	if err != nil {
		return reflectionServiceServer{}, fmt.Errorf("unable to create deliver descriptor: %w", err)
	}
	desc := &AppDescriptor{
		Chain:         chainDescriptor,
		Codec:         codecDescriptor,
		Configuration: configurationDescriptor,
		QueryServices: queryServiceDescriptor,
		Tx:            txDescriptor,
	}
	ifaceList := make([]string, len(desc.Codec.Interfaces))
	ifaceImplementers := make(map[string][]string, len(desc.Codec.Interfaces))
	for i, iface := range desc.Codec.Interfaces {
		ifaceList[i] = iface.Fullname
		impls := make([]string, len(iface.InterfaceImplementers))
		for j, impl := range iface.InterfaceImplementers {
			impls[j] = impl.TypeUrl
		}
		ifaceImplementers[iface.Fullname] = impls
	}
	return reflectionServiceServer{
		desc:                  desc,
		interfacesList:        ifaceList,
		interfaceImplementers: ifaceImplementers,
	}, nil
}

func (r reflectionServiceServer) GetAppDescriptor(_ context.Context, _ *GetAppDescriptorRequest) (*GetAppDescriptorResponse, error) {
	return &GetAppDescriptorResponse{App: r.desc}, nil
}

func (r reflectionServiceServer) ListAllInterfaces(_ context.Context, _ *ListAllInterfacesRequest) (*ListAllInterfacesResponse, error) {
	return &ListAllInterfacesResponse{InterfaceNames: r.interfacesList}, nil
}

func (r reflectionServiceServer) ListImplementations(_ context.Context, request *ListImplementationsRequest) (*ListImplementationsResponse, error) {
	implementers, ok := r.interfaceImplementers[request.InterfaceName]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "interface name %s does not exist", request.InterfaceName)
	}

	return &ListImplementationsResponse{ImplementationMessageNames: implementers}, nil
}

// newCodecDescriptor describes the codec given the codectypes.InterfaceRegistry
func newCodecDescriptor(ir codectypes.InterfaceRegistry) (*CodecDescriptor, error) {
	registeredInterfaces := ir.ListAllInterfaces()
	interfaceDescriptors := make([]*InterfaceDescriptor, len(registeredInterfaces))

	for i, iface := range registeredInterfaces {
		implementers := ir.ListImplementations(iface)
		interfaceImplementers := make([]*InterfaceImplementerDescriptor, len(implementers))
		for j, implementer := range implementers {
			pb, err := ir.Resolve(implementer)
			if err != nil {
				return nil, fmt.Errorf("unable to resolve implementing type %s for interface %s", implementer, iface)
			}
			pbName := proto.MessageName(pb)
			if pbName == "" {
				return nil, fmt.Errorf("unable to get proto name for implementing type %s for interface %s", implementer, iface)
			}
			interfaceImplementers[j] = &InterfaceImplementerDescriptor{
				Fullname: pbName,
				TypeUrl:  implementer,
			}
		}
		interfaceDescriptors[i] = &InterfaceDescriptor{
			Fullname: iface,
			// NOTE(fdymylja): this could be filled, but it won't be filled as of now
			// doing this would require us to fully rebuild in a (dependency) transitive way the proto
			// registry of the supported proto.Messages for the application, this could be easily
			// done if we weren't relying on gogoproto which does not allow us to iterate over the
			// registry. Achieving this right now would mean to start slowly building descriptors
			// getting their files dependencies, building those dependencies then rebuilding the
			// descriptor builder. It's too much work as of now.
			InterfaceAcceptingMessages: nil,
			InterfaceImplementers:      interfaceImplementers,
		}
	}

	return &CodecDescriptor{
		Interfaces: interfaceDescriptors,
	}, nil
}

func newQueryServiceDescriptor(srv *grpc.Server) *QueryServicesDescriptor {
	svcInfo := srv.GetServiceInfo()
	queryServices := make([]*QueryServiceDescriptor, 0, len(svcInfo))
	for name, info := range svcInfo {
		methods := make([]*QueryMethodDescriptor, len(info.Methods))
		for i, svcMethod := range info.Methods {
			methods[i] = &QueryMethodDescriptor{
				Name:          svcMethod.Name,
				FullQueryPath: fmt.Sprintf("/%s/%s", name, svcMethod.Name),
			}
		}
		queryServices = append(queryServices, &QueryServiceDescriptor{
			Fullname: name,
			Methods:  methods,
		})
	}
	return &QueryServicesDescriptor{QueryServices: queryServices}
}

func newTxDescriptor(ir codectypes.InterfaceRegistry, signingModes []string) (*TxDescriptor, error) {
	// get base tx type name
	txPbName := proto.MessageName(&tx.Tx{})
	if txPbName == "" {
		return nil, fmt.Errorf("unable to get *tx.Tx protobuf name")
	}
	// get msgs
	msgImplementers := ir.ListImplementations(sdk.MsgInterfaceProtoName)
	svcMsgImplementers := ir.ListImplementations(sdk.ServiceMsgInterfaceProtoName)

	msgsDesc := make([]*MsgDescriptor, 0, len(msgImplementers)+len(svcMsgImplementers))

	// process sdk.ServiceMsg
	for _, svcMsg := range svcMsgImplementers {
		resolved, err := ir.Resolve(svcMsg)
		if err != nil {
			return nil, fmt.Errorf("unable to resolve sdk.ServiceMsg %s: %w", svcMsg, err)
		}
		pbName := proto.MessageName(resolved)
		if pbName == "" {
			return nil, fmt.Errorf("unable to get proto name for sdk.ServiceMsg %s", svcMsg)
		}

		msgsDesc = append(msgsDesc, &MsgDescriptor{Msg: &MsgDescriptor_ServiceMsg{
			ServiceMsg: &ServiceMsgDescriptor{
				RequestFullname: pbName,
				RequestRoute:    svcMsg,
				RequestTypeUrl:  svcMsg,
				// NOTE(fdymylja): this cannot be filled as of now, the Configurator is not held inside the *BaseApp type
				// but is local to specific applications, hence we have no way of getting the MsgServer's descriptors
				// which contain response information.
				ResponseFullname: "",
			},
		}})
	}

	signModesDesc := make([]*SigningModeDescriptor, len(signingModes))
	for i, m := range signingModes {
		signModesDesc[i] = &SigningModeDescriptor{
			Name: m,
			// NOTE(fdymylja): this cannot be filled as of now, auth and the sdk itself don't support as of now
			// a service which allows to get authentication metadata for the provided sign mode.
			AuthnInfoProviderMethodFullname: "",
		}
	}
	return &TxDescriptor{
		Fullname: txPbName,
		Authn: &AuthnDescriptor{
			SignModes: signModesDesc,
		},
		Msgs: msgsDesc,
	}, nil
}
