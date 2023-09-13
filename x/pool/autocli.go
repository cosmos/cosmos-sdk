package pool

import (
	"fmt"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	poolv1 "cosmossdk.io/api/cosmos/pool/v1"

	"github.com/cosmos/cosmos-sdk/version"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: poolv1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "CommunityPool",
					Use:       "community-pool",
					Short:     "Query the amount of coins in the community pool",
					Example:   fmt.Sprintf(`$ %s query pool community-pool`, version.AppName),
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: poolv1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod:      "FundCommunityPool",
					Use:            "fund-community-pool [amount]",
					Short:          "Funds the community pool with the specified amount",
					Example:        fmt.Sprintf(`$ %s tx pool fund-community-pool 100uatom --from mykey`, version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "amount"}},
				},
			},
		},
	}
}
