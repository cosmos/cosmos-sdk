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
		Use:   "grant [granter] [grantee] [limit] or grant [granter] [grantee] [limit] [period] [periodLimit]",
		Short: "Grant Fee allowance to an address",
		Long: strings.TrimSpace(
			fmt.Sprintf(
				`Grant authorization to pay fees from your address. Note, the'--from' flag is
				ignored as it is implied from [granter].

Examples:
%s tx %s grant cosmos1skjw... cosmos1skjw... 100stake --expiration 36000 or
%s tx %s grant cosmos1skjw... cosmos1skjw... 100stake --period 3600 --periodlimit 10stake --expiration 36000
				`, version.AppName, types.ModuleName, version.AppName, types.ModuleName,
			),
		),
		Args: cobra.ExactArgs(3),
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

			limit, err := sdk.ParseCoinsNormalized(args[2])
			if err != nil {
				return err
			}

			exp, err := cmd.Flags().GetInt64(FlagExpiration)
			if err != nil {
				return err
			}

			expDuration := time.Duration(exp) * time.Second

			basic := types.BasicFeeAllowance{
				SpendLimit: limit,
				Expiration: types.ExpiresAtTime(time.Now().Add(expDuration)),
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

					if periodClock > exp {
						return fmt.Errorf("period(%d) cannot be greater than the expiration(%d)", periodClock, exp)
					}

					periodic := types.PeriodicFeeAllowance{
						Basic:            basic,
						Period:           types.ClockDuration(time.Duration(periodClock) * time.Second),
						PeriodReset:      types.ExpiresAtTime(time.Now().Add(time.Duration(periodClock) * time.Second)),
						PeriodSpendLimit: periodLimit,
						PeriodCanSpend:   periodLimit,
					}

					grant = &periodic

				} else {
					return fmt.Errorf("invalid number of args %d", len(args))
				}
			}

			msg, err := types.NewMsgGrantFeeAllowance(grant, granter, grantee)
			if err != nil {
				return err
			}

			svcMsgClientConn := &msgservice.ServiceMsgClientConn{}
			feeGrantMsgClient := types.NewMsgClient(svcMsgClientConn)
			_, err = feeGrantMsgClient.GrantFeeAllowance(cmd.Context(), msg)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), svcMsgClientConn.GetMsgs()...)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().Int64(FlagExpiration, int64(365*24*60*60), "The second unit of time duration which the grant is active for the user; Default is a year")
	cmd.Flags().Int64(FlagPeriod, 0, "period specifies the time duration in which period_spend_limit coins can be spent before that allowance is reset")
	cmd.Flags().String(FlagPeriodLimit, "", "// period limit specifies the maximum number of coins that can be spent in the period")

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
			feeGrantMsgClient := types.NewMsgClient(svcMsgClientConn)
			_, err = feeGrantMsgClient.RevokeFeeAllowance(cmd.Context(), &msg)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), svcMsgClientConn.GetMsgs()...)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
