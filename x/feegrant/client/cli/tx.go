package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/feegrant/types"
)

// flag for feegrant module
const (
	FlagExpiration  = "expiration"
	FlagPeriod      = "period"
	FlagPeriodLimit = "period-limit"
	FlagSpendLimit  = "spend-limit"
	FlagAllowedMsgs = "allowed-messages"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	feegrantTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Feegrant transactions subcommands",
		Long:                       "Grant and revoke fee allowance for a grantee by a granter",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	feegrantTxCmd.AddCommand(
		NewCmdFeeGrant(),
		NewCmdRevokeFeegrant(),
	)

	return feegrantTxCmd
}

// NewCmdFeeGrant returns a CLI command handler for creating a MsgGrantFeeAllowance transaction.
func NewCmdFeeGrant() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grant [granter] [grantee]",
		Short: "Grant Fee allowance to an address",
		Long: strings.TrimSpace(
			fmt.Sprintf(
				`Grant authorization to pay fees from your address. Note, the'--from' flag is
				ignored as it is implied from [granter].

Examples:
%s tx %s grant cosmos1skjw... cosmos1skjw... --spend-limit 100stake --expiration 36000 or
%s tx %s grant cosmos1skjw... cosmos1skjw... --spend-limit 100stake --period 3600 --period-limit 10stake --expiration 36000 or
%s tx %s grant cosmos1skjw... cosmos1skjw... --spend-limit 100stake --expiration 36000 
	--allowed-messages "/cosmos.gov.v1beta1.Msg/SubmitProposal,/cosmos.gov.v1beta1.Msg/Vote"
				`, version.AppName, types.ModuleName, version.AppName, types.ModuleName, version.AppName, types.ModuleName,
			),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			cmd.Flags().Set(flags.FlagFrom, args[0])
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			grantee, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			granter := clientCtx.GetFromAddress()
			sl, err := cmd.Flags().GetString(FlagSpendLimit)
			if err != nil {
				return err
			}

			// if `FlagSpendLimit` isn't set, limit will be nil
			limit, err := sdk.ParseCoinsNormalized(sl)
			if err != nil {
				return err
			}

			exp, err := cmd.Flags().GetInt64(FlagExpiration)
			if err != nil {
				return err
			}

			basic := types.BasicFeeAllowance{
				SpendLimit: limit,
			}

			if exp != 0 {
				expDuration := time.Now().Add(time.Duration(exp) * time.Second)
				basic.Expiration = &expDuration
			}

			var grant types.FeeAllowanceI
			grant = &basic

			periodClock, err := cmd.Flags().GetInt64(FlagPeriod)
			if err != nil {
				return err
			}

			periodLimitVal, err := cmd.Flags().GetString(FlagPeriodLimit)
			if err != nil {
				return err
			}

			// Check any of period or periodLimit flags set, If set consider it as periodic fee allowance.
			if periodClock > 0 || periodLimitVal != "" {
				periodLimit, err := sdk.ParseCoinsNormalized(periodLimitVal)
				if err != nil {
					return err
				}

				if periodClock > 0 && periodLimit != nil {
					if exp > 0 && periodClock > exp {
						return fmt.Errorf("period(%d) cannot be greater than the expiration(%d)", periodClock, exp)
					}

					periodic := types.PeriodicFeeAllowance{
						Basic:            basic,
						Period:           time.Duration(periodClock) * time.Second,
						PeriodReset:      time.Now().Add(time.Duration(periodClock) * time.Second),
						PeriodSpendLimit: periodLimit,
						PeriodCanSpend:   periodLimit,
					}

					grant = &periodic

				} else {
					return fmt.Errorf("invalid number of args %d", len(args))
				}
			}

			allowedMsgs, err := cmd.Flags().GetStringSlice(FlagAllowedMsgs)
			if err != nil {
				return err
			}

			if len(allowedMsgs) > 0 {
				grant, err = types.NewAllowedMsgFeeAllowance(grant, allowedMsgs)
				if err != nil {
					return err
				}
			}

			msg, err := types.NewMsgGrantFeeAllowance(grant, granter, grantee)
			if err != nil {
				return err
			}

			svcMsgClientConn := &msgservice.ServiceMsgClientConn{}
			msgClient := types.NewMsgClient(svcMsgClientConn)
			_, err = msgClient.GrantFeeAllowance(cmd.Context(), msg)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), svcMsgClientConn.GetMsgs()...)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().StringSlice(FlagAllowedMsgs, []string{}, "Set of allowed messages for fee allowance")
	cmd.Flags().Int64(FlagExpiration, 0, "The second unit of time duration which the grant is active for the user")
	cmd.Flags().String(FlagSpendLimit, "", "Spend limit specifies the max limit can be used, if not mentioned there is no limit")
	cmd.Flags().Int64(FlagPeriod, 0, "period specifies the time duration in which period_spend_limit coins can be spent before that allowance is reset")
	cmd.Flags().String(FlagPeriodLimit, "", "period limit specifies the maximum number of coins that can be spent in the period")

	return cmd
}

// NewCmdRevokeFeegrant returns a CLI command handler for creating a MsgRevokeFeeAllowance transaction.
func NewCmdRevokeFeegrant() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke [granter_address] [grantee_address]",
		Short: "revoke fee-grant",
		Long: strings.TrimSpace(
			fmt.Sprintf(`revoke fee grant from a granter to a grantee. Note, the'--from' flag is
			ignored as it is implied from [granter_address].

Example:
 $ %s tx %s revoke cosmos1skj.. cosmos1skj..
			`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Flags().Set(flags.FlagFrom, args[0])
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			grantee, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgRevokeFeeAllowance(clientCtx.GetFromAddress(), grantee)
			svcMsgClientConn := &msgservice.ServiceMsgClientConn{}
			msgClient := types.NewMsgClient(svcMsgClientConn)
			_, err = msgClient.RevokeFeeAllowance(cmd.Context(), &msg)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), svcMsgClientConn.GetMsgs()...)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
