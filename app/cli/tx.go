package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
)

type TxCommand struct{ *cobra.Command }

func (TxCommand) IsAutoGroupType() {}

func ProvideTxCommand(commands []TxCommand) RootCommand {
	cmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	for _, c := range commands {
		if c.Command != nil {
			cmd.AddCommand(c.Command)
		}
	}

	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return RootCommand{cmd}
}
