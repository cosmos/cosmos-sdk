package circuit

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	circuitv1 "cosmossdk.io/api/cosmos/circuit/v1"
	"fmt"
	"github.com/cosmos/cosmos-sdk/version"
)

func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: circuitv1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "DisabledList",
					Use:       "disabled-list",
					Short:     "Query for all disabled message types",
				},
				{
					RpcMethod:      "Account",
					Use:            "account [address]",
					Short:          "Query for account permissions",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "address"}},
				},
				{
					RpcMethod: "Accounts",
					Use:       "accounts",
					Short:     "Query for all account permissions",
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: circuitv1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "AuthorizeCircuitBreaker",
					Use:       "authorize [grantee] [permission_level] [limit_type_urls] --from [granter]",
					Short:     "Authorize an account to trip the circuit breaker.",
					Long: `Authorize an account to trip the circuit breaker.
		"SOME_MSGS" =     1,
		"ALL_MSGS" =      2,
		"SUPER_ADMIN" =   3,`,
					Example: fmt.Sprintf(`%s circuit authorize [address] 0 "cosmos.bank.v1beta1.MsgSend,cosmos.bank.v1beta1.MsgMultiSend"`, version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "grantee"},
						// TODO: permission_level are an embeded type
						{ProtoField: "permission_level"},
						{ProtoField: "limit_type_urls"},
					},
					FlagOptions: map[string]*autocliv1.FlagOptions{
						"granter": {
							Name:  "from",
							Usage: "--from [granter]",
						},
					},
				},
				{
					RpcMethod: "TripCircuitBreaker",
					Use:       "disable [type_url]",
					Short:     "disable a message from being executed",
					Example:   fmt.Sprintf(`%s circuit disable "cosmos.bank.v1beta1.MsgSend,cosmos.bank.v1beta1.MsgMultiSend"`, version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						// Client ctx contains authority
						{ProtoField: "type_url", Varargs: true},
					},
				},
				{
					RpcMethod: "ResetCircuitBreaker",
					Use:       "reset [type_url]",
					Short:     "Enable a message to be executed",
					Example:   fmt.Sprintf(`%s circuit reset "cosmos.bank.v1beta1.MsgSend,cosmos.bank.v1beta1.MsgMultiSend"`, version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						// Client ctx contains authority
						{ProtoField: "type_url", Varargs: true},
					},
				},
			},
		},
	}
}
