package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"cosmossdk.io/core/address"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

var FlagSplit = "split"

// NewTxCmd returns a root CLI command handler for all x/bank transaction commands.
func NewTxCmd(ac address.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Bank transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewSendTxCmd(ac),
		NewMultiSendTxCmd(ac),
	)

	return txCmd
}

// NewSendTxCmd returns a CLI command handler for creating a MsgSend transaction.
func NewSendTxCmd(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send [from_key_or_address] [to_address] [amount]",
		Short: "Send funds from one account to another.",
		Long: `Send funds from one account to another.
Note, the '--from' flag is ignored as it is implied from [from_key_or_address].
When using '--dry-run' a key name cannot be used, only a bech32 address.
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			toAddr, err := ac.StringToBytes(args[1])
			if err != nil {
				return err
			}

			coins, err := sdk.ParseCoinsNormalized(args[2])
			if err != nil {
				return err
			}

			if len(coins) == 0 {
				return fmt.Errorf("invalid coins")
			}

			msg := types.NewMsgSend(clientCtx.GetFromAddress(), toAddr, coins)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewMultiSendTxCmd returns a CLI command handler for creating a MsgMultiSend transaction.
// For a better UX this command is limited to send funds from one account to two or more accounts.
func NewMultiSendTxCmd(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "multi-send [from_key_or_address] [to_address_1 to_address_2 ...] [amount]",
		Short: "Deprecated: multi-send is deprecated as of Cosmos SDK v0.53",
		Args:  cobra.MinimumNArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("deprecated: multi-send is deprecated as of Cosmos SDK v0.53")
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
