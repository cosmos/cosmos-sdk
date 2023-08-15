package rpc

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	cmtv1beta1 "cosmossdk.io/api/cosmos/base/tendermint/v1beta1"
)

func NewCometBFTCommands() *cometModule {
	return &cometModule{}
}

type cometModule struct{}

func (m cometModule) IsOnePerModuleType() {}
func (m cometModule) IsAppModule()        {}

func (m cometModule) Name() string {
	return "comet"
}

func (m cometModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
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
					Use:            "block-by-height [height]",
					Short:          "Query for a committed block by height",
					Long:           "Query for a specific committed block using the CometBFT RPC `block_by_height` method",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "height"}},
				},
				{
					RpcMethod: "GetLatestValidatorSet",
					Use:       "validator-set-latest",
					Short:     "Query for the latest validator set",
				},
				{
					RpcMethod:      "GetValidatorSetByHeight",
					Use:            "validator-set-by-height [height]",
					Short:          "Query for a validator set by height",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "height"}},
				},
				{
					RpcMethod: "ABCIQuery",
					Use:       "abci-query [path] [data] [height] <prove>",
					Short:     "Query the ABCI application for data",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "path"},
						{ProtoField: "data"},
						{ProtoField: "height"},
						{ProtoField: "prove", Optional: true},
					},
				},
			},
		}}
}
