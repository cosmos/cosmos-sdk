package cli

import (
	"cosmossdk.io/x/benchmark"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"
)

func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        "benchmark",
		Short:                      "benchmark transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewLoadTestCmd(),
	)

	return txCmd
}

func NewLoadTestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "load-test [from_key_or_address]",
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &benchmark.MsgLoadTest{Caller: clientCtx.FromAddress}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
