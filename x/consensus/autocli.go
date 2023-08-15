package consensus

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	cmtv1beta1 "cosmossdk.io/api/cosmos/base/tendermint/v1beta1"
	consensusv1 "cosmossdk.io/api/cosmos/consensus/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: consensusv1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query the current consensus parameters",
				},
			},
			SubCommands: map[string]*autocliv1.ServiceCommandDescriptor{
				"comet": {
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
				},
			},
		},
		// Tx is purposely left empty, as the only tx is MsgUpdateParams which is gov gated.
	}
}
