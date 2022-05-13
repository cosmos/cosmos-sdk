package cli

import (
	"errors"
	"fmt"
	feegrantcli "github.com/cosmos/cosmos-sdk/x/feegrant/client/cli"

	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	"github.com/cosmos/cosmos-sdk/x/authz"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Flag names and values
const (
	FlagSpendLimit        = "spend-limit"
	FlagMsgType           = "msg-type"
	FlagExpiration        = "expiration"
	FlagAllowedValidators = "allowed-validators"
	FlagDenyValidators    = "deny-validators"
	FlagPeriod            = "period"
	FlagPeriodLimit       = "period-limit"
	delegate              = "delegate"
	redelegate            = "redelegate"
	unbond                = "unbond"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	AuthorizationTxCmd := &cobra.Command{
		Use:                        authz.ModuleName,
		Short:                      "Authorization transactions subcommands",
		Long:                       "Authorize and revoke access to execute transactions on behalf of your address",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	AuthorizationTxCmd.AddCommand(
		NewCmdGrantAuthorization(),
		NewCmdRevokeAuthorization(),
		NewCmdExecAuthorization(),
	)

	return AuthorizationTxCmd
}

func NewCmdGrantAuthorization() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "grant <grantee> <authorization_type=\"send\"|\"generic\"|\"delegate\"|\"unbond\"|\"redelegate\"> --from <granter>",
		Aliases:    nil,
		SuggestFor: nil,
		Short:      "Grant authorization to an address",
		Long: strings.TrimSpace(
			fmt.Sprintf(`create a new grant authorization to an address to execute a transaction on your behalf:

Examples:
 $ %s tx %s grant cosmos1skjw.. send %s --spend-limit=1000stake --from=cosmos1skl..
 $ %s tx %s grant cosmos1skjw.. generic --msg-type=/cosmos.gov.v1.MsgVote --from=cosmos1sk..
	`, version.AppName, authz.ModuleName, bank.SendAuthorization{}.MsgTypeURL(), version.AppName, authz.ModuleName),
		),
		Example:                "",
		ValidArgs:              nil,
		ValidArgsFunction:      nil,
		Args:                   cobra.ExactArgs(2),
		ArgAliases:             nil,
		BashCompletionFunction: "",
		Deprecated:             "",
		Annotations:            nil,
		Version:                "",
		PersistentPreRun:       nil,
		PersistentPreRunE:      nil,
		PreRun:                 nil,
		PreRunE:                nil,
		Run:                    nil,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			grantee, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			expire, err := getExpireTime(cmd)
			if err != nil {
				return err
			}

			limit, err := cmd.Flags().GetString(FlagSpendLimit)
			if err != nil {
				return err
			}

			var spendLimit sdk.Coins
			if limit != "" {
				spendLimit, err = sdk.ParseCoinsNormalized(limit)
				if err != nil {
					return err
				}
				if !spendLimit.IsAllPositive() {
					return fmt.Errorf("spend-limit should be greater than zero")
				}
			}

			var authorization authz.Authorization
			switch args[1] {
			case "send":
				periodClock, err := cmd.Flags().GetInt64(FlagPeriod)
				if err != nil {
					return err
				}

				periodLimitVal, err := cmd.Flags().GetString(FlagPeriodLimit)
				if err != nil {
					return err
				}

				// Check any of period or periodLimit flags set, If set consider it as periodic authorization.
				if periodClock > 0 || periodLimitVal != "" {
					periodLimit, err := sdk.ParseCoinsNormalized(periodLimitVal)
					if err != nil {
						return err
					}

					if periodClock <= 0 {
						return fmt.Errorf("period clock was not set")
					}

					if periodLimit == nil {
						return fmt.Errorf("period limit was not set")
					}
					periodReset := feegrantcli.GetPeriodReset(periodClock)
					if expire != nil && periodReset.Sub(*expire) > 0 {
						return fmt.Errorf("period (%d) cannot reset after expiration (%v)", periodClock, expire)
					}

					authorization = bank.NewPeriodicSendAuthorization(
						spendLimit,
						expire,
						feegrantcli.GetPeriod(periodClock),
						periodLimit,
						feegrantcli.GetPeriodReset(periodClock),
					)

				} else {
					if expire != nil {
						return fmt.Errorf("expire can only be used with periodic authorizations")
					}
					authorization = bank.NewSendAuthorization(spendLimit)
				}

			case "generic":
				msgType, err := cmd.Flags().GetString(FlagMsgType)
				if err != nil {
					return err
				}

				authorization = authz.NewGenericAuthorization(msgType)
			case delegate, unbond, redelegate:
				limit, err := cmd.Flags().GetString(FlagSpendLimit)
				if err != nil {
					return err
				}

				allowValidators, err := cmd.Flags().GetStringSlice(FlagAllowedValidators)
				if err != nil {
					return err
				}

				denyValidators, err := cmd.Flags().GetStringSlice(FlagDenyValidators)
				if err != nil {
					return err
				}

				var delegateLimit *sdk.Coin
				if limit != "" {
					spendLimit, err := sdk.ParseCoinsNormalized(limit)
					if err != nil {
						return err
					}

					if !spendLimit.IsAllPositive() {
						return fmt.Errorf("spend-limit should be greater than zero")
					}
					delegateLimit = &spendLimit[0]
				}

				allowed, err := bech32toValidatorAddresses(allowValidators)
				if err != nil {
					return err
				}

				denied, err := bech32toValidatorAddresses(denyValidators)
				if err != nil {
					return err
				}

				switch args[1] {
				case delegate:
					authorization, err = staking.NewStakeAuthorization(allowed, denied, staking.AuthorizationType_AUTHORIZATION_TYPE_DELEGATE, delegateLimit)
				case unbond:
					authorization, err = staking.NewStakeAuthorization(allowed, denied, staking.AuthorizationType_AUTHORIZATION_TYPE_UNDELEGATE, delegateLimit)
				default:
					authorization, err = staking.NewStakeAuthorization(allowed, denied, staking.AuthorizationType_AUTHORIZATION_TYPE_REDELEGATE, delegateLimit)
				}
				if err != nil {
					return err
				}

			default:
				return fmt.Errorf("invalid authorization type, %s", args[1])
			}

			msg, err := authz.NewMsgGrant(clientCtx.GetFromAddress(), grantee, authorization, expire)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
		PostRun:                    nil,
		PostRunE:                   nil,
		PersistentPostRun:          nil,
		PersistentPostRunE:         nil,
		FParseErrWhitelist:         cobra.FParseErrWhitelist{},
		CompletionOptions:          cobra.CompletionOptions{},
		TraverseChildren:           false,
		Hidden:                     false,
		SilenceErrors:              false,
		SilenceUsage:               false,
		DisableFlagParsing:         false,
		DisableAutoGenTag:          false,
		DisableFlagsInUseLine:      false,
		DisableSuggestions:         false,
		SuggestionsMinimumDistance: 0,
	}
	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String(FlagMsgType, "", "The Msg method name for which we are creating a GenericAuthorization")
	cmd.Flags().String(FlagSpendLimit, "", "SpendLimit for Send Authorization, an array of Coins allowed spend")
	cmd.Flags().StringSlice(FlagAllowedValidators, []string{}, "Allowed validators addresses separated by ,")
	cmd.Flags().StringSlice(FlagDenyValidators, []string{}, "Deny validators addresses separated by ,")
	cmd.Flags().Int64(FlagExpiration, 0, "Expire time as Unix timestamp. Set zero (0) for no expiry. Default is 0.")
	cmd.Flags().Int64(FlagPeriod, 0, "period specifies the time duration(in seconds) in which period_limit coins can be spent before that allowance is reset (ex: 3600)")
	cmd.Flags().String(FlagPeriodLimit, "", "period limit specifies the maximum number of coins that can be spent in the period")
	return cmd
}

func getExpireTime(cmd *cobra.Command) (*time.Time, error) {
	exp, err := cmd.Flags().GetInt64(FlagExpiration)
	if err != nil {
		return nil, err
	}
	if exp == 0 {
		return nil, nil
	}
	e := time.Unix(exp, 0)
	return &e, nil
}

func NewCmdRevokeAuthorization() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke [grantee] [msg-type-url] --from=[granter]",
		Short: "revoke authorization",
		Long: strings.TrimSpace(
			fmt.Sprintf(`revoke authorization from a granter to a grantee:
Example:
 $ %s tx %s revoke cosmos1skj.. %s --from=cosmos1skj..
			`, version.AppName, authz.ModuleName, bank.SendAuthorization{}.MsgTypeURL()),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			grantee, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			granter := clientCtx.GetFromAddress()
			msgAuthorized := args[1]
			msg := authz.NewMsgRevoke(granter, grantee, msgAuthorized)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func NewCmdExecAuthorization() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec [tx-json-file] --from [grantee]",
		Short: "execute tx on behalf of granter account",
		Long: strings.TrimSpace(
			fmt.Sprintf(`execute tx on behalf of granter account:
Example:
 $ %s tx %s exec tx.json --from grantee
 $ %s tx bank send <granter> <recipient> --from <granter> --chain-id <chain-id> --generate-only > tx.json && %s tx %s exec tx.json --from grantee
			`, version.AppName, authz.ModuleName, version.AppName, version.AppName, authz.ModuleName),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			grantee := clientCtx.GetFromAddress()

			if offline, _ := cmd.Flags().GetBool(flags.FlagOffline); offline {
				return errors.New("cannot broadcast tx during offline mode")
			}

			theTx, err := authclient.ReadTxFromFile(clientCtx, args[0])
			if err != nil {
				return err
			}
			msg := authz.NewMsgExec(grantee, theTx.GetMsgs())

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func bech32toValidatorAddresses(validators []string) ([]sdk.ValAddress, error) {
	vals := make([]sdk.ValAddress, len(validators))
	for i, validator := range validators {
		addr, err := sdk.ValAddressFromBech32(validator)
		if err != nil {
			return nil, err
		}
		vals[i] = addr
	}
	return vals, nil
}
