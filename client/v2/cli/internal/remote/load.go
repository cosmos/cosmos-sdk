package remote

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoregistry"
)

type ChainInfo struct {
	ModuleOptions     map[string]*autocliv1.ModuleOptions
	GRPCClient        *grpc.ClientConn
	FileDescriptorSet *protoregistry.Files
}

func LoadChainInfo(config *ChainConfig) (*ChainInfo, error) {
	var client *grpc.ClientConn
	for _, endpoint := range config.TrustedGRPCEndpoints {
		var err error
		client, err = grpc.Dial(endpoint)
		if err != nil {
			return nil, err
		}
	}

	// TODO get module options

	return &ChainInfo{
		ModuleOptions: nil,
		GRPCClient:    client,
	}, nil
}
