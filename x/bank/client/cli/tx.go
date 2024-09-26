package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/x/bank/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
)

var FlagSplit = "split"

// NewTxCmd returns a root CLI command handler for all x/bank transaction commands.
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Bank transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewMultiSendTxCmd(),
	)

	return txCmd
}

// NewMultiSendTxCmd returns a CLI command handler for creating a MsgMultiSend transaction.
// This version allows sending funds from one account to one or more accounts.
func NewMultiSendTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "multi-send <from_key_or_address> <to_address_1>... <amount>",
		Short: "Send funds from one account to one or more accounts.",
		Long: `Send funds from one account to one or more accounts.
By default, sends the [amount] to each address in the list.
Using the '--split' flag, the [amount] is split equally between the addresses.
Note, the '--from' flag is ignored as it is implied from [from_key_or_address] and
separate addresses with space.
When using '--dry-run' a key name cannot be used, only a bech32 address.`,
		Example: fmt.Sprintf("%s tx bank multi-send cosmos1... cosmos1... cosmos1... 10stake", version.AppName),
		Args:    cobra.MinimumNArgs(3), // Changed minimum argument count to 3
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0]) // Set the first argument as the sender
			if err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			coins, err := sdk.ParseCoinsNormalized(args[len(args)-1]) // The last argument is the amount
			if err != nil {
				return err
			}

			if coins.IsZero() {
				return errors.New("must send positive amount")
			}

			split, err := cmd.Flags().GetBool(FlagSplit)
			if err != nil {
				return err
			}

			totalAddrs := sdkmath.NewInt(int64(len(args) - 2)) // Calculate the number of recipients
			// coins to be received by the addresses
			sendCoins := coins
			if split {
				sendCoins = coins.QuoInt(totalAddrs) // Logic to split the amount among recipients
			}

			var output []types.Output
			for _, arg := range args[1 : len(args)-1] { // Process each recipient
				_, err = clientCtx.AddressCodec.StringToBytes(arg)
				if err != nil {
					return err
				}

				output = append(output, types.NewOutput(arg, sendCoins))
			}

			// Calculate the total amount to be sent by the sender
			var amount sdk.Coins
			if split {
				amount = sendCoins.MulInt(totalAddrs)
			} else {
				amount = coins.MulInt(totalAddrs)
			}

			fromAddr, err := clientCtx.AddressCodec.BytesToString(clientCtx.FromAddress)
			if err != nil {
				return err
			}

			msg := types.NewMsgMultiSend(types.NewInput(fromAddr, amount), output)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Bool(FlagSplit, false, "Send the equally split token amount to each address")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
