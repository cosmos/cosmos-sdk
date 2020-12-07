package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/feegrant/types"
)

// flag for feegrant module
const (
	FlagExpiration = "expiration"
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

func NewCmdFeeGrant() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grant [grantee] [limit] --from [granter]",
		Short: "Grant Fee allowance to an address",
		Long: strings.TrimSpace(
			fmt.Sprintf(
				`Grant authorization to use fee from your address.

Examples:
%s tx %s grant cosmos1skjw... 1000stake --from=cosmos1skjw...
				`, version.AppName, types.ModuleName,
			),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			grantee, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			limit, err := sdk.ParseCoinsNormalized(args[2])
			if err != nil {
				return err
			}

			period := time.Duration(viper.GetInt64(FlagExpiration)) * time.Second
			_ = period // TODO

			basic := types.BasicFeeAllowance{
				SpendLimit: limit,
				// TODO
				// Expiration: period,
			}

			msg, err := types.NewMsgGrantFeeAllowance(&basic, clientCtx.GetFromAddress(), grantee)
			if err != nil {
				return err
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().Int64(FlagExpiration, int64(365*24*60*60), "The second unit of time duration which the grant is active for the user; Default is a year")
	return cmd
}

func NewCmdRevokeFeegrant() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke [grantee_address] --from=[granter_address]",
		Short: "revoke fee-grant",
		Long: strings.TrimSpace(
			fmt.Sprintf(`revoke fee grant from a granter to a grantee:

Example:
 $ %s tx %s revoke cosmos1skj.. --from=cosmos1skj..
			`, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			grantee, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			granter := clientCtx.GetFromAddress()

			msg := types.MsgRevokeFeeAllowance{
				Granter: granter,
				Grantee: grantee,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
