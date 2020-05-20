package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// NewTxCmd returns a root CLI command handler for all x/bank transaction commands.
func NewTxCmd(cliCtx context.CLIContext) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Bank transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(NewSendTxCmd(cliCtx))

	return txCmd
}

// NewSendTxCmd returns a CLI command handler for creating a MsgSend transaction.
func NewSendTxCmd(cliCtx context.CLIContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send [from_key_or_address] [to_address] [amount]",
		Short: "Create and/or sign and broadcast a MsgSend transaction",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx = cliCtx.InitWithInputAndFrom(cmd.InOrStdin(), args[0])

			toAddr, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			coins, err := sdk.ParseCoins(args[2])
			if err != nil {
				return err
			}

			msg := types.NewMsgSend(cliCtx.GetFromAddress(), toAddr, coins)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTx(cliCtx, msg)
		},
	}

	return flags.PostCommands(cmd)[0]
}
