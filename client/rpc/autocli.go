package rpc

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	cmtv1beta1 "cosmossdk.io/api/cosmos/base/tendermint/v1beta1"
)

func NewCometModule() *cometModule {
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
