package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

// Transaction command flags
const (
	FlagDelayed = "delayed"
)

// GetTxCmd returns vesting module's transaction commands.
func GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Vesting transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewMsgCreateVestingAccountCmd(),
		NewMsgCreatePeriodicVestingAccountCmd(),
	)

	return txCmd
}

// NewMsgCreateVestingAccountCmd returns a CLI command handler for creating a
// MsgCreateVestingAccount transaction.
func NewMsgCreateVestingAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-vesting-account [to_address] [amount] [end_time]",
		Short: "Create a new vesting account funded with an allocation of tokens.",
		Long: `Create a new vesting account funded with an allocation of tokens. The
account can either be a delayed or continuous vesting account, which is determined
by the '--delayed' flag. All vesting accouts created will have their start time
set by the committed block's time. The end_time must be provided as a UNIX epoch
timestamp.`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			toAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return err
			}

			endTime, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				return err
			}

			delayed, _ := cmd.Flags().GetBool(FlagDelayed)

			msg := types.NewMsgCreateVestingAccount(clientCtx.GetFromAddress(), toAddr, amount, endTime, delayed)
			svcMsgClientConn := &msgservice.ServiceMsgClientConn{}
			msgClient := types.NewMsgClient(svcMsgClientConn)
			_, err = msgClient.CreateVestingAccount(cmd.Context(), msg)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), svcMsgClientConn.GetMsgs()...)
		},
	}

	cmd.Flags().Bool(FlagDelayed, false, "Create a delayed vesting account if true")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

type InputPeriod struct {
	Coins  string `json:"coins"`
	Length int64  `json:"length_seconds"`
}

// NewMsgCreateVestingAccountCmd returns a CLI command handler for creating a
// MsgCreateVestingAccount transaction.
func NewMsgCreatePeriodicVestingAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-periodic-vesting-account [to_address] [periods_json_file]",
		Short: "Create a new vesting account funded with an allocation of tokens.",
		Long: `Create a new vesting account funded with an allocation of tokens. This takes a destination address and 
		a period json file. 
		Where periods.json contains:

		An array of coin strings and unix epoch times for coins to vest

[
{
  "coins": "10test",
  "length_seconds":2592000 //30 days
 },
 {
	"coins": "10test",
	"length_seconds":2592000 //30 days
   },
]
		`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			toAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			contents, err := ioutil.ReadFile(args[1])
			if err != nil {
				return err
			}

			var inputPeriods []InputPeriod

			err = json.Unmarshal(contents, inputPeriods)
			if err != nil {
				return err
			}

			var periods []types.Period

			for i, p := range inputPeriods {

				amount, err := sdk.ParseCoinsNormalized(p.Coins)
				if err != nil {
					return err
				}

				if p.Length < 1 {
					return fmt.Errorf("invalid period length of %d in period %d, length must be greater than 0", p.Length, i)
				}
				period := types.Period{Length: p.Length, Amount: amount}
				periods = append(periods, period)
			}

			msg := types.NewMsgCreatePeriodicVestingAccount(clientCtx.GetFromAddress(), toAddr, periods)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Bool(FlagDelayed, false, "Create a delayed vesting account if true")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
