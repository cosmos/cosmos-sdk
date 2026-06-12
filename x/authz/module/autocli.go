package authz

import (
	"fmt"

	authzv1beta1 "cosmossdk.io/api/cosmos/authz/v1beta1"
	autocli "cosmossdk.io/core/autocli"

	"github.com/cosmos/cosmos-sdk/version"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocli.ModuleOptions {
	return &autocli.ModuleOptions{
		Query: &autocli.ServiceCommandDescriptor{
			Service: authzv1beta1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocli.RpcCommandOptions{
				{
					RpcMethod: "Grants",
					Use:       "grants [granter-addr] [grantee-addr] <msg-type-url>",
					Short:     "Query grants for a granter-grantee pair and optionally a msg-type-url",
					Long:      "Query authorization grants for a granter-grantee pair. If msg-type-url is set, it will select grants only for that msg type.",
					Example:   fmt.Sprintf("%s query authz grants cosmos1skj.. cosmos1skjwj.. %s", version.AppName, bank.SendAuthorization{}.MsgTypeURL()),
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "granter"},
						{ProtoField: "grantee"},
						{ProtoField: "msg_type_url", Optional: true},
					},
				},
				{
					RpcMethod: "GranterGrants",
					Use:       "grants-by-granter [granter-addr]",
					Short:     "Query authorization grants granted by granter",
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "granter"},
					},
				},
				{
					RpcMethod: "GranteeGrants",
					Use:       "grants-by-grantee [grantee-addr]",
					Short:     "Query authorization grants granted to a grantee",
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "grantee"},
					},
				},
			},
		},
		Tx: &autocli.ServiceCommandDescriptor{
			Service:              authzv1beta1.Msg_ServiceDesc.ServiceName,
			EnhanceCustomCommand: false, // use custom commands only until v0.51
			RpcCommandOptions: []*autocli.RpcCommandOptions{
				{
					RpcMethod: "Exec",
					Use:       "exec [msg-json-file] --from [grantee]",
					Short:     "Execute tx on behalf of granter account",
					Example:   fmt.Sprintf("$ %s tx authz exec msg.json --from grantee\n $ %[1]s tx bank send [granter] [recipient] [amount] --generate-only | jq .body.messages > msg.json && %[1]s tx authz exec msg.json --from grantee", version.AppName),
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "msgs", Varargs: true},
					},
				},
				{
					RpcMethod: "Revoke",
					Use:       "revoke [grantee] [msg-type-url] --from [granter]",
					Short:     `Revoke authorization from a granter to a grantee`,
					Example: fmt.Sprintf(`%s tx authz revoke cosmos1skj.. %s --from=cosmos1skj..`,
						version.AppName, bank.SendAuthorization{}.MsgTypeURL()),
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "grantee"},
						{ProtoField: "msg_type_url"},
					},
				},
			},
		},
	}
}
