package cmtservice

import (
	cmtv1beta1 "cosmossdk.io/api/cosmos/base/tendermint/v1beta1"
	autocli "cosmossdk.io/core/autocli"
)

var CometBFTAutoCLIDescriptor = &autocli.ServiceCommandDescriptor{
	Service: cmtv1beta1.Service_ServiceDesc.ServiceName,
	RpcCommandOptions: []*autocli.RpcCommandOptions{
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
			Use:            "block-by-height [height]",
			Short:          "Query for a committed block by height",
			Long:           "Query for a specific committed block using the CometBFT RPC `block_by_height` method",
			PositionalArgs: []*autocli.PositionalArgDescriptor{{ProtoField: "height"}},
		},
		{
			RpcMethod: "GetLatestValidatorSet",
			Use:       "validator-set",
			Alias:     []string{"validator-set-latest", "comet-validator-set", "cometbft-validator-set", "tendermint-validator-set"},
			Short:     "Query for the latest validator set",
		},
		{
			RpcMethod:      "GetValidatorSetByHeight",
			Use:            "validator-set-by-height [height]",
			Short:          "Query for a validator set by height",
			PositionalArgs: []*autocli.PositionalArgDescriptor{{ProtoField: "height"}},
		},
		{
			RpcMethod: "ABCIQuery",
			Skip:      true,
		},
		{
			RpcMethod: "GetLatestBlockResults",
			Use:       "block-results-latest",
			Short:     "Query for the latest block results",
		},
		{
			RpcMethod:      "GetBlockResults",
			Use:            "block-results [height]",
			Short:          "Query for block results by height",
			PositionalArgs: []*autocli.PositionalArgDescriptor{{ProtoField: "height"}},
		},
	},
}

// NewCometBFTCommands is a fake `appmodule.Module` to be considered as a module
// and be added in AutoCLI.
func NewCometBFTCommands() *cometModule {
	return &cometModule{}
}

type cometModule struct{}

func (m cometModule) IsOnePerModuleType() {}
func (m cometModule) IsAppModule()        {}

func (m cometModule) Name() string {
	return "comet"
}

func (m cometModule) AutoCLIOptions() *autocli.ModuleOptions {
	return &autocli.ModuleOptions{
		Query: CometBFTAutoCLIDescriptor,
	}
}
