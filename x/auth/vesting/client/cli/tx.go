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
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

// Transaction command flags
const (
	FlagDelayed = "delayed"
	FlagDest    = "dest"
	FlagLockup  = "lockup"
	FlagMerge   = "merge"
	FlagVesting = "vesting"
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
		NewMsgCreateClawbackVestingAccountCmd(),
		NewMsgClawbackCmd(),
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

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Bool(FlagDelayed, false, "Create a delayed vesting account if true")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

type VestingData struct {
	StartTime int64         `json:"start_time"`
	Periods   []InputPeriod `json:"periods"`
}

type InputPeriod struct {
	Coins  string `json:"coins"`
	Length int64  `json:"length"`
}

// readScheduleFile reads the file at path and unmarshals it to get the schedule.
// Returns start time, periods, and error.
func readScheduleFile(path string) (int64, []types.Period, error) {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return 0, nil, err
	}

	var data VestingData
	if err := json.Unmarshal(contents, &data); err != nil {
		return 0, nil, err
	}

	startTime := data.StartTime
	periods := make([]types.Period, len(data.Periods))

	for i, p := range data.Periods {
		amount, err := sdk.ParseCoinsNormalized(p.Coins)
		if err != nil {
			return 0, nil, err
		}
		if p.Length < 1 {
			return 0, nil, fmt.Errorf("invalid period length of %d in period %d, length must be greater than 0", p.Length, i)
		}

		periods[i] = types.Period{Length: p.Length, Amount: amount}
	}

	return startTime, periods, nil
}

// NewMsgCreateClawbackVestingAccountCmd returns a CLI command handler for creating a
// MsgCreateClawbackVestingAccount transaction.
func NewMsgCreateClawbackVestingAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-clawback-vesting-account [to_address]",
		Short: "Create a new vesting account funded with an allocation of tokens, subject to clawback.",
		Long: `Must provide a lockup periods file (--lockup), a vesting periods file (--vesting), or both.
If both files are given, they must describe schedules for the same total amount.
If one file is omitted, it will default to a schedule that immediately unlocks or vests the entire amount.
The described amount of coins will be transferred from the --from address to the vesting account.
Unvested coins may be "clawed back" by the funder with the clawback command.
Coins may not be transferred out of the account if they are locked or unvested, but may be staked.
Staking rewards are subject to a proportional vesting encumbrance.

A periods file is a JSON object describing a sequence of unlocking or vesting events,
with a start time and an array of coins strings and durations relative to the start or previous event.`,
		Example: `Sample period file contents:
{
  "start_time": 1625204910,
  "periods": [
    {
      "coins": "10test",
      "length": 2592000 // 30 days
    },
    {
      "coins": "10test",
      "length": 2592000 // 30 days
    }
  ]
}
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			toAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			lockupFile, _ := cmd.Flags().GetString(FlagLockup)
			vestingFile, _ := cmd.Flags().GetString(FlagVesting)
			if lockupFile == "" && vestingFile == "" {
				return fmt.Errorf("must specify at least one of %s or %s", FlagLockup, FlagVesting)
			}

			var (
				lockupStart, vestingStart     int64
				lockupPeriods, vestingPeriods []types.Period
			)
			if lockupFile != "" {
				lockupStart, lockupPeriods, err = readScheduleFile(lockupFile)
				if err != nil {
					return err
				}
			}
			if vestingFile != "" {
				vestingStart, vestingPeriods, err = readScheduleFile(vestingFile)
				if err != nil {
					return err
				}
			}

			commonStart, _ := types.AlignSchedules(lockupStart, vestingStart, lockupPeriods, vestingPeriods)

			merge, _ := cmd.Flags().GetBool(FlagMerge)

			msg := types.NewMsgCreateClawbackVestingAccount(clientCtx.GetFromAddress(), toAddr, commonStart, lockupPeriods, vestingPeriods, merge)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Bool(FlagMerge, false, "Merge new amount and schedule with existing ClawbackVestingAccount, if any")
	cmd.Flags().String(FlagLockup, "", "Path to file containing unlocking periods")
	cmd.Flags().String(FlagVesting, "", "Path to file containing vesting periods")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewMsgClawbackCmd returns a CLI command handler for creating a
// MsgClawback transaction.
func NewMsgClawbackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clawback [address]",
		Short: "Transfer unvested amount out of a ClawbackVestingAccount.",
		Long: `Must be requested by the original funder address (--from).
		May provide a destination address (--dest), otherwise the coins return to the funder.
		Delegated or undelegating staking tokens will be transferred in the delegated (undelegating) state.
		The recipient is vulnerable to slashing, and must act to unbond the tokens if desired.
		`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			var dest sdk.AccAddress
			destString, _ := cmd.Flags().GetString(FlagDest)
			if destString != "" {
				dest, err = sdk.AccAddressFromBech32(destString)
				if err != nil {
					return fmt.Errorf("invalid destination address: %w", err)
				}
			}

			msg := types.NewMsgClawback(clientCtx.GetFromAddress(), addr, dest)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagDest, "", "Address of destination (defaults to funder)")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
