package safety

import (
	"fmt"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	safetyv1 "cosmossdk.io/api/cosmos/safety/v1"

	"github.com/cosmos/cosmos-sdk/version"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: safetyv1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query the x/safety params: authority, operator and operator scope",
				},
				{
					RpcMethod: "DisabledList",
					Use:       "disabled-list",
					Short:     "Query the disabled patterns",
				},
				{
					RpcMethod:      "IsAllowed",
					Use:            "is-allowed [msg-type-url]",
					Short:          "Query whether a msg type URL may execute",
					Example:        fmt.Sprintf(`%s query safety is-allowed /cosmos.bank.v1beta1.MsgSend`, version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "msg_type_url"}},
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: safetyv1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Trip",
					Use:       "trip [pattern] [pattern...] --from [operator]",
					Short:     "Disable msgs matching the given patterns",
					Long: `Disable msgs matching the given patterns. Only the operator may trip, and only
patterns covered by its scope. A pattern is "*", a prefix such as /cosmos.bank.*,
or an exact msg type URL. Pass each pattern as its own argument.`,
					Example: fmt.Sprintf(`%[1]s tx safety trip /cosmos.bank.v1beta1.MsgSend /cosmos.bank.v1beta1.MsgMultiSend --from operator
%[1]s tx safety trip '/cosmos.staking.*' --from operator`, version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "patterns", Varargs: true},
					},
				},
				{
					RpcMethod: "Reset",
					Use:       "reset [pattern] [pattern...] --from [sender]",
					Short:     "Re-enable msgs by removing patterns from the disabled list",
					Long: `Re-enable msgs by removing patterns from the disabled list. The authority may
reset any pattern; the operator may reset patterns covered by its scope. Each
pattern must match a disabled entry exactly. Pass each pattern as its own argument.`,
					Example: fmt.Sprintf(`%[1]s tx safety reset /cosmos.bank.v1beta1.MsgSend /cosmos.bank.v1beta1.MsgMultiSend --from operator
%[1]s tx safety reset '/cosmos.staking.*' --from operator`, version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "patterns", Varargs: true},
					},
				},
				{
					RpcMethod: "UpdateParams",
					Use:       "update-params-proposal [params]",
					Short:     "Submit a proposal to update the x/safety params. Note: the entire params must be provided.",
					Example:   fmt.Sprintf(`%s tx safety update-params-proposal '{ "authority": "cosmos1...", "operator": "cosmos1...", "scope": ["/cosmos.bank.*"] }' --title "Name the safety operator" --summary "Grant the operator /cosmos.bank.* scope" --deposit 10000000stake --from proposer`, version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "params"},
					},
					GovProposal: true,
				},
			},
		},
	}
}
