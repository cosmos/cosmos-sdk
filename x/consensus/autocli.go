package consensus

import (
	"fmt"

	consensusv1 "cosmossdk.io/api/cosmos/consensus/v1"
	autocli "cosmossdk.io/core/autocli"

	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	"github.com/cosmos/cosmos-sdk/version"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocli.ModuleOptions {
	return &autocli.ModuleOptions{
		Query: &autocli.ServiceCommandDescriptor{
			Service: consensusv1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocli.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query the current consensus parameters",
				},
			},
			SubCommands: map[string]*autocli.ServiceCommandDescriptor{
				"comet": cmtservice.CometBFTAutoCLIDescriptor,
			},
		},
		Tx: &autocli.ServiceCommandDescriptor{
			Service: consensusv1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocli.RpcCommandOptions{
				{
					RpcMethod: "UpdateParams",
					Use:       "update-params-proposal [params]",
					Short:     "Submit a proposal to update consensus module params. Note: the entire params must be provided.",
					Example:   fmt.Sprintf(`%s tx consensus update-params-proposal '{ params }'`, version.AppName),
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "block"},
						{ProtoField: "evidence"},
						{ProtoField: "validator"},
						{ProtoField: "abci"},
					},
					GovProposal: true,
				},
			},
		},
	}
}
