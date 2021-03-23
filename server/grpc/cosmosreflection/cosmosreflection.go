package cosmosreflection

import (
	"context"
	"encoding/json"
	"fmt"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc"
	"log"
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
	desc *AppDescriptor
}

func newReflectionServiceServer(grpcSrv *grpc.Server, conf Config) (reflectionServiceServer, error) {
	// set chain descriptor
	chainDescriptor := &ChainDescriptor{Id: conf.ChainID}
	// set configuration descriptor
	configurationDescriptor := &ConfigurationDescriptor{
		Bech32AccountAddressPrefix:     conf.SdkConfig.GetBech32AccountAddrPrefix(),
		Bech32ValidatorAddressPrefix:   conf.SdkConfig.GetBech32ValidatorAddrPrefix(),
		Bech32ConsensusAddressPrefix:   conf.SdkConfig.GetBech32ConsensusAddrPrefix(),
		Bech32AccountPublicKeyPrefix:   conf.SdkConfig.GetBech32AccountPubPrefix(),
		Bech32ValidatorPublicKeyPrefix: conf.SdkConfig.GetBech32ValidatorPubPrefix(),
		Bech32ConsensusPublicKeyPrefix: conf.SdkConfig.GetBech32ConsensusPubPrefix(),
		Purpose:                        conf.SdkConfig.GetPurpose(),
		CoinType:                       conf.SdkConfig.GetCoinType(),
		FullFundraiserPath:             conf.SdkConfig.GetFullFundraiserPath(),
		FullBip44Path:                  conf.SdkConfig.GetFullBIP44Path(),
	}
	// set codec descriptor
	codecDescriptor, err := newCodecDescriptor(conf.InterfaceRegistry)
	if err != nil {
		return reflectionServiceServer{}, fmt.Errorf("unable to create codec descriptor: %w", err)
	}
	// set query service descriptor
	queryServiceDescriptor, err := newQueryServiceDescriptor(grpcSrv)
	if err != nil {
		return reflectionServiceServer{}, fmt.Errorf("unable to create query services descriptor: %w", err)
	}
	// set deliver descriptor
	deliverDescriptor, err := newDeliverDescriptor(conf.InterfaceRegistry, conf.SigningModes)
	if err != nil {
		return reflectionServiceServer{}, fmt.Errorf("unable to create deliver descriptor: %w", err)
	}
	desc := &AppDescriptor{
		Chain:         chainDescriptor,
		Codec:         codecDescriptor,
		Configuration: configurationDescriptor,
		QueryServices: queryServiceDescriptor,
		Tx:            deliverDescriptor,
	}
	b, _ := json.Marshal(desc)
	log.Printf("%s", b)
	return reflectionServiceServer{desc: desc}, nil
}

func (r reflectionServiceServer) GetAppDescriptor(ctx context.Context, request *GetAppDescriptorRequest) (*GetAppDescriptorResponse, error) {
	panic("implement me")
}

func (r reflectionServiceServer) ListAllInterfaces(ctx context.Context, request *ListAllInterfacesRequest) (*ListAllInterfacesResponse, error) {
	panic("implement me")
}

func (r reflectionServiceServer) ListImplementations(ctx context.Context, request *ListImplementationsRequest) (*ListImplementationsResponse, error) {
	panic("implement me")
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
			Fullname:                iface,
			InterfaceAcceptingTypes: nil, // NOTE(fdymylja): this will be used in the future when we will replace *anypb.Any fields with the interface
			InterfaceImplementers:   interfaceImplementers,
		}
	}

	return &CodecDescriptor{
		Interfaces: interfaceDescriptors,
	}, nil
}

func newQueryServiceDescriptor(srv *grpc.Server) (*QueryServicesDescriptor, error) {
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
	return &QueryServicesDescriptor{QueryServices: queryServices}, nil
}

func newDeliverDescriptor(ir codectypes.InterfaceRegistry, signingModes []string) (*TxDescriptor, error) {
	// get base tx type name
	txPbName := proto.MessageName(&tx.Tx{})
	if txPbName == "" {
		return nil, fmt.Errorf("unable to get *tx.Tx protobuf name")
	}
	// get msgs
	msgImplementers := ir.ListImplementations(sdk.MsgInterfaceProtoName)
	svcMsgImplementers := ir.ListImplementations(sdk.ServiceMsgInterfaceProtoName)

	msgsDesc := make([]*MsgDescriptor, 0, len(msgImplementers)+len(svcMsgImplementers))

	// process sdk.Msg
	for _, msg := range msgImplementers {
		pb, err := ir.Resolve(msg)
		if err != nil {
			return nil, fmt.Errorf("unable to resolve sdk.Msg %s: %w", msg, err)
		}
		pbName := proto.MessageName(pb)
		if pbName == "" {
			return nil, fmt.Errorf("unable to get proto name for sdk.Msg %s", msg)
		}
		msgsDesc = append(msgsDesc, &MsgDescriptor{Msg: &MsgDescriptor_LegacyMsg{
			LegacyMsg: &LegacyMsgDescriptor{
				Fullname: pbName,
				TypeUrl:  msg,
			},
		}})
	}
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
				Fullname: pbName,
				Route:    svcMsg,
				TypeUrl:  svcMsg,
			},
		}})
	}

	signModesDesc := make([]*SigningModeDescriptor, len(signingModes))
	for i, m := range signingModes {
		signModesDesc[i] = &SigningModeDescriptor{Name: m}
	}
	return &TxDescriptor{
		Fullname:   txPbName,
		AuthConfig: &AuthConfigDescriptor{SigningModes: signModesDesc},
		Msgs:       msgsDesc,
	}, nil
}
