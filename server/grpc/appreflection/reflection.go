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
			Fullname:                   iface,
			InterfaceAcceptingMessages: nil, // NOTE(fdymylja): this will be used in the future when we will replace *anypb.Any fields with the interface
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
			},
		}})
	}

	signModesDesc := make([]*SigningModeDescriptor, len(signingModes))
	for i, m := range signingModes {
		signModesDesc[i] = &SigningModeDescriptor{
			Name:                            m,
			AuthnInfoProviderMethodFullname: "", // this cannot be filled as of now
		}
	}
	return &TxDescriptor{
		Fullname: txPbName,
		Authn: &AuthnDescriptor{
			SignModes: signModesDesc,
		}, // TODO
		Msgs: msgsDesc,
	}, nil
}
