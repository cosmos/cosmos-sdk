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

func (m cometModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: cmtv1beta1.Service_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "GetNodeInfo",
				},
				{
					RpcMethod: "GetSyncing",
				},
				{
					RpcMethod: "GetLatestBlock",
				},
				{
					RpcMethod: "GetBlockByHeight",
					Use:   "block-by-height [height]",
					Short: "Query for a committed block by height",
					Long:  "Query for a specific committed block using the CometBFT RPC `block_by_height` method",		
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "height"}},
				},
				{
					RpcMethod: "GetLatestValidatorSet",
				},
				{
					RpcMethod: "GetValidatorSetByHeight",
				},
				{
					RpcMethod: "ABCIQuery",
				},
			},
		}}
}
