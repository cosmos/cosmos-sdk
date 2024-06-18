package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"cosmossdk.io/x/bank/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
)

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
// For a better UX this command is limited to send funds from one account to two or more accounts.
func NewMultiSendTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "multi-send [from_key_or_address] [to_address_1:amount_1;to_address_2:amount_2 ...]",
		Short: "Send funds from one account to two or more accounts.",
		Long: `Send funds from one account to two or more accounts.
Note, the '--from' flag is ignored as it is implied from [from_key_or_address] and 
separate output with ';', separate address between coins with ':', separate coins with ','.
When using '--dry-run' a key name cannot be used, only a bech32 address.`,
		Example: fmt.Sprintf("%s tx bank multi-send node0 cosmos1shlek07kh379w6hye8lxgy3nx58spnkhkjl8fp:10stake,100testtoken;cosmos1fqjltn3l7nxme663wz69g2zyd7ejs0gq4alrkn:1stake,10testtoken", version.AppName),
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			// amount to be send from the from address
			var amount sdk.Coins
			var outputs []types.Output

			outputList := strings.Split(args[1], ";") // for each output
			for _, outputStr := range outputList {
				output := strings.Split(outputStr, ":") // address between coins
				if len(output) != 2 {
					return fmt.Errorf("invalid output %s", output)
				}

				_, err := clientCtx.AddressCodec.StringToBytes(output[0])
				if err != nil {
					return fmt.Errorf("invalid bech32 string %s", output[0])
				}

				coins, err := sdk.ParseCoinsNormalized(output[1])
				if err != nil {
					return err
				}
				if coins.Len() == 0 {
					return fmt.Errorf("must send positive amount %s", coins.String())
				}

				coins = coins.Sort()

				amount = amount.Add(coins...)

				outputs = append(outputs, types.NewOutput(output[0], coins))
			}

			fromAddr, err := clientCtx.AddressCodec.BytesToString(clientCtx.FromAddress)
			if err != nil {
				return err
			}

			msg := types.NewMsgMultiSend(types.NewInput(fromAddr, amount), outputs)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
