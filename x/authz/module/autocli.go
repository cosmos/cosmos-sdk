package authz

import (
	"fmt"
	"strings"

	authzv1beta1 "cosmossdk.io/api/cosmos/authz/v1beta1"
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/authz"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
)

const (
	FlagSpendLimit        = "spend-limit"
	FlagMsgType           = "msg-type"
	FlagExpiration        = "expiration"
	FlagAllowedValidators = "allowed-validators"
	FlagDenyValidators    = "deny-validators"
	FlagAllowList         = "allow-list"
	delegate              = "delegate"
	redelegate            = "redelegate"
	unbond                = "unbond"
)

func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: authzv1beta1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Grants",
					Use:       "grants [granter-addr] [grantee-addr] [msg-type-url]?",
					Short:     "query grants for a granter-grantee pair and optionally a msg-type-url",
					Long: strings.TrimSpace(
						fmt.Sprintf(`Query authorization grants for a granter-grantee pair. If msg-type-url
is set, it will select grants only for that msg type.
Examples:
$ %s query %s grants cosmos1skj.. cosmos1skjwj..
$ %s query %s grants cosmos1skjw.. cosmos1skjwj.. %s
`,
							version.AppName, authz.ModuleName,
							version.AppName, authz.ModuleName, bank.SendAuthorization{}.MsgTypeURL()),
					),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "granter"},
						{ProtoField: "grantee"},
						// TODO: make this optional
						{ProtoField: "msg_type_url"},
					},
				},
				{
					RpcMethod: "GranterGrants",
					Use:       "grants-by-granter [granter-addr]",
					Short:     "query authorization grants granted by granter",
					Long: strings.TrimSpace(
						fmt.Sprintf(`Query authorization grants granted by granter.
Examples:
$ %s q %s grants-by-granter cosmos1skj..
`,
							version.AppName, authz.ModuleName),
					),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "granter"},
					},
				},
				{
					RpcMethod: "GranteeGrants",
					Use:       "grants-by-grantee [grantee-addr]",
					Short:     "query authorization grants granted to a grantee",
					Long: strings.TrimSpace(
						fmt.Sprintf(`Query authorization grants granted to a grantee.
Examples:
$ %s q %s grants-by-grantee cosmos1skj..
`,
							version.AppName, authz.ModuleName),
					),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "grantee"},
					},
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: authzv1beta1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Grant",
					Use:       "grant <grantee> <authorization_type=\"send\"|\"generic\"|\"delegate\"|\"unbond\"|\"redelegate\"> --from <granter>",
					Short:     "Grant authorization to an address",
					Long: strings.TrimSpace(
						fmt.Sprintf(`create a new grant authorization to an address to execute a transaction on your behalf:

Examples:
 $ %s tx %s grant cosmos1skjw.. send --spend-limit=1000stake --from=cosmos1skl..
 $ %s tx %s grant cosmos1skjw.. generic --msg-type=/cosmos.gov.v1.MsgVote --from=cosmos1sk..
	`, version.AppName, authz.ModuleName, version.AppName, authz.ModuleName),
					),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "grantee"},
						{ProtoField: "grant"},
					},
				},
			},
		},
	}
}
