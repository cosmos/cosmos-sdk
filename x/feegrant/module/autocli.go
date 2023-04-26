package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	modulev1 "cosmossdk.io/api/cosmos/feegrant/v1beta1"
	"cosmossdk.io/x/feegrant"
	"fmt"
	"github.com/cosmos/cosmos-sdk/version"
	"strings"
)

const (
	FlagExpiration  = "expiration"
	FlagPeriod      = "period"
	FlagPeriodLimit = "period-limit"
	FlagSpendLimit  = "spend-limit"
	FlagAllowedMsgs = "allowed-messages"
)

func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: modulev1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Grant",
					Use:       "grant [granter] [grantee]",
					Short:     "Query for a grant",
					Long: strings.TrimSpace(
						fmt.Sprintf(`Query details for a grant. 
You can find the fee-grant of a granter and grantee.

Example:
$ %s query feegrant grant [granter] [grantee]
`, version.AppName),
					),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "granter"}, {ProtoField: "grantee"}},
				},
				{
					RpcMethod: "GrantsByGrantee",
					Use:       "grants-by-grantee [grantee]",
					Short:     "Query all grants of a grantee",
					Long: strings.TrimSpace(
						fmt.Sprintf(`Queries all the grants for a grantee address.

Example:
$ %s query feegrant grants-by-grantee [grantee]
`, version.AppName),
					),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "grantee"}},
				},
				{
					RpcMethod: "GrantsByGranter",
					Use:       "grants-by-granter [granter]",
					Short:     "Query all grants of a granter",
					Long: strings.TrimSpace(
						fmt.Sprintf(`Queries all the grants issued for a granter address.

Example:
$ %s query feegrant grants-by-granter [granter]
`, version.AppName),
					),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "granter"}},
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: modulev1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Grant",
					Use:       "grant [granter_key_or_address] [grantee]",
					Short:     "Grant Fee allowance to an address",
					Long: strings.TrimSpace(
						fmt.Sprintf(
							`Grant authorization to pay fees from your address. Note, the'--from' flag is
				ignored as it is implied from [granter].

Examples:
%s tx %s grant cosmos1skjw... cosmos1skjw... --spend-limit 100stake --expiration 2022-01-30T15:04:05Z or
%s tx %s grant cosmos1skjw... cosmos1skjw... --spend-limit 100stake --period 3600 --period-limit 10stake --expiration 2022-01-30T15:04:05Z or
%s tx %s grant cosmos1skjw... cosmos1skjw... --spend-limit 100stake --expiration 2022-01-30T15:04:05Z 
	--allowed-messages "/cosmos.gov.v1beta1.MsgSubmitProposal,/cosmos.gov.v1beta1.MsgVote"
				`, version.AppName, feegrant.ModuleName, version.AppName, feegrant.ModuleName, version.AppName, feegrant.ModuleName,
						),
					),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "granter"}, {ProtoField: "grantee"}},
					FlagOptions: map[string]*autocliv1.FlagOptions{
						FlagExpiration: {
							Name:  FlagExpiration,
							Usage: "The RFC 3339 timestamp after which the grant expires for the user",
						},
						FlagAllowedMsgs: {
							Name:  FlagAllowedMsgs,
							Usage: "Set of allowed messages for fee allowance",
						},
						FlagSpendLimit: {
							Name:  FlagSpendLimit,
							Usage: "Spend limit specifies the max limit can be used, if not mentioned there is no limit",
						},
						FlagPeriod: {
							Name:  FlagPeriod,
							Usage: "Period specifies the time duration(in seconds) in which period_limit coins can be spent before that allowance is reset (ex: 3600)",
						},
						FlagPeriodLimit: {
							Name:  FlagPeriodLimit,
							Usage: "Period limit specifies the maximum number of coins that can be spent in the period",
						},
					},
				},
				{
					RpcMethod: "Revoke",
					Use:       "revoke [granter_key_or_address] [grantee]",
					Short:     "Revoke Fee allowance from an address",
					Long: strings.TrimSpace(
						fmt.Sprintf(`revoke fee grant from a granter to a grantee. Note, the'--from' flag is
			ignored as it is implied from [granter].

Example:
 $ %s tx %s revoke cosmos1skj.. cosmos1skj..
			`, version.AppName, feegrant.ModuleName),
					),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "granter"}, {ProtoField: "grantee"}},
				},
			},
		},
	}
}
