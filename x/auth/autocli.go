package auth

import (
	"fmt"

	authv1beta1 "cosmossdk.io/api/cosmos/auth/v1beta1"
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"github.com/cosmos/cosmos-sdk/version"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: authv1beta1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod:      "Account",
					Use:            "account [address]",
					Short:          "query account by address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "address"}},
				},
				{
					RpcMethod:      "AccountAddressByID",
					Use:            "address-by-acc-num [acc-num]",
					Short:          "query account address by account number",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},
				{
					RpcMethod: "ModuleAccounts",
					Use:       "module-accounts",
					Short:     "Query all module accounts",
				},
				{
					RpcMethod:      "ModuleAccountByName",
					Use:            "module-account [module-name]",
					Short:          "Query module account info by module name",
					Example:        fmt.Sprintf("%s q auth module-account gov", version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "name"}},
				},
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query the current auth parameters",
				},
			},
		},
		// Tx is purposely left empty, as the only tx is MsgUpdateParams which is gov gated.
	}
}
