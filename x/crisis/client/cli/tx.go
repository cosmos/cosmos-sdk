package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/x/crisis/types"
)

// NewTxCmd returns a root CLI command handler for all x/crisis transaction commands.
func NewTxCmd(ctx context.CLIContext) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Crisis transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(NewMsgVerifyInvariantTxCmd(ctx))

	return txCmd
}

// NewMsgVerifyInvariantTxCmd returns a CLI command handler for creating a
// MsgVerifyInvariant transaction.
func NewMsgVerifyInvariantTxCmd(ctx context.CLIContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invariant-broken [module-name] [invariant-route]",
		Short: "Submit proof that an invariant broken to halt the chain",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := ctx.InitWithInput(cmd.InOrStdin())

			senderAddr := cliCtx.GetFromAddress()
			moduleName, route := args[0], args[1]

			msg := types.NewMsgVerifyInvariant(senderAddr, moduleName, route)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTx(cliCtx, msg)
		},
	}

	return flags.PostCommands(cmd)[0]
}
