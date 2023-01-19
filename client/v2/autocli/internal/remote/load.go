package remote

import (
	"context"
	"crypto/tls"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

type ChainInfo struct {
	ModuleOptions     map[string]*autocliv1.ModuleOptions
	GRPCClient        *grpc.ClientConn
	FileDescriptorSet *protoregistry.Files
	Context           context.Context
}

func LoadChainInfo(chain string, config *ChainConfig, reload bool) (*ChainInfo, error) {
	var client *grpc.ClientConn
	for _, endpoint := range config.TrustedGRPCEndpoints {
		var err error
		var creds credentials.TransportCredentials
		if endpoint.Insecure {
			creds = insecure.NewCredentials()
		} else {
			creds = credentials.NewTLS(&tls.Config{
				MinVersion: tls.VersionTLS12,
			})
		}
		client, err = grpc.Dial(endpoint.Endpoint, grpc.WithTransportCredentials(creds))
		if err != nil {
			return nil, err
		}
	}

	ctx := context.Background()

	autocliQueryClient := autocliv1.NewQueryClient(client)
	apppOptionsRes, err := autocliQueryClient.AppOptions(ctx, &autocliv1.AppOptionsRequest{})
	if err != nil {
		return nil, err
	}

	reflectionClient := reflectionv1.NewReflectionServiceClient(client)
	fdRes, err := reflectionClient.FileDescriptors(ctx, &reflectionv1.FileDescriptorsRequest{})
	if err != nil {
		return nil, err
	}

	files, err := protodesc.NewFiles(&descriptorpb.FileDescriptorSet{File: fdRes.Files})
	if err != nil {
		return nil, err
	}

	return &ChainInfo{
		ModuleOptions:     apppOptionsRes.ModuleOptions,
		GRPCClient:        client,
		Context:           ctx,
		FileDescriptorSet: files,
	}, nil
}
