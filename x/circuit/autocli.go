package circuit

import (
	"fmt"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	circuitv1 "cosmossdk.io/api/cosmos/circuit/v1"

	"github.com/cosmos/cosmos-sdk/version"
)

func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: circuitv1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod:      "Account",
					Use:            "account [address]",
					Short:          "Query a specific account's permissions",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "address"}},
				},
				{
					RpcMethod: "Accounts",
					Use:       "accounts",
					Short:     "Query all account permissions",
				},
				{
					RpcMethod: "DisabledList",
					Use:       "disabled-list",
					Short:     "Query a list of all disabled message types",
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: circuitv1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "AuthorizeCircuitBreaker",
					Use:       "authorize [grantee] [permissions_json] --from [granter]",
					Short:     "Authorize an account to trip the circuit breaker.",
					Long: `Authorize an account to trip the circuit breaker.
"SOME_MSGS" =     1,
"ALL_MSGS" =      2,
"SUPER_ADMIN" =   3,`,
					Example: fmt.Sprintf(`%s circuit authorize [address] '{"level":1,"limit_type_urls":["cosmos.bank.v1beta1.MsgSend,cosmos.bank.v1beta1.MsgMultiSend"]}'"`, version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "grantee"},
						{ProtoField: "permissions"}, // TODO(@julienrbrt) Support flattening msg for setting each field as a positional arg
					},
				},
				{
					RpcMethod: "TripCircuitBreaker",
					Use:       "disable [msg_type_urls]",
					Short:     "Disable a message from being executed",
					Example:   fmt.Sprintf(`%s circuit disable "cosmos.bank.v1beta1.MsgSend cosmos.bank.v1beta1.MsgMultiSend"`, version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "msg_type_urls", Varargs: true},
					},
				},
				{
					RpcMethod: "ResetCircuitBreaker",
					Use:       "reset [msg_type_urls]",
					Short:     "Enable a message to be executed",
					Example:   fmt.Sprintf(`%s circuit reset "cosmos.bank.v1beta1.MsgSend cosmos.bank.v1beta1.MsgMultiSend"`, version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "msg_type_urls", Varargs: true},
					},
				},
			},
		},
	}
}
