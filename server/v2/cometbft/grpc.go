package cometbft

import (
	"google.golang.org/grpc"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	cmtv1beta1 "cosmossdk.io/api/cosmos/base/tendermint/v1beta1"
	"cosmossdk.io/core/address"
	"cosmossdk.io/server/v2/cometbft/client/rpc"

	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
)

// GRPCServiceRegistrar returns a function that registers the CometBFT gRPC service
func (c *Consensus[T]) GRPCServiceRegistrar(
	cometRPC rpc.CometRPC,
	consensusAddressCodec address.ConsensusAddressCodec,
) func(srv *grpc.Server) error {
	return func(srv *grpc.Server) error {
		cmtservice.RegisterServiceServer(srv, cmtservice.NewQueryServer(cometRPC, c.Query, consensusAddressCodec))
		return nil
	}
}

// CometBFTAutoCLIDescriptor is the auto-generated CLI descriptor for the CometBFT service
var CometBFTAutoCLIDescriptor = &autocliv1.ServiceCommandDescriptor{
	Service: cmtv1beta1.Service_ServiceDesc.ServiceName,
	RpcCommandOptions: []*autocliv1.RpcCommandOptions{
		{
			RpcMethod: "GetNodeInfo",
			Use:       "node-info",
			Short:     "Query the current node info",
		},
		{
			RpcMethod: "GetSyncing",
			Use:       "syncing",
			Short:     "Query node syncing status",
		},
		{
			RpcMethod: "GetLatestBlock",
			Use:       "block-latest",
			Short:     "Query for the latest committed block",
		},
		{
			RpcMethod:      "GetBlockByHeight",
			Use:            "block-by-height <height>",
			Short:          "Query for a committed block by height",
			Long:           "Query for a specific committed block using the CometBFT RPC `block_by_height` method",
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "height"}},
		},
		{
			RpcMethod: "GetLatestValidatorSet",
			Use:       "validator-set",
			Alias:     []string{"validator-set-latest", "comet-validator-set", "cometbft-validator-set", "tendermint-validator-set"},
			Short:     "Query for the latest validator set",
		},
		{
			RpcMethod:      "GetValidatorSetByHeight",
			Use:            "validator-set-by-height <height>",
			Short:          "Query for a validator set by height",
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "height"}},
		},
		{
			RpcMethod: "ABCIQuery",
			Skip:      true,
		},
	},
}
