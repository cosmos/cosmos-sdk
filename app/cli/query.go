package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/spf13/cobra"
)

type QueryCommand struct{ *cobra.Command }

func (QueryCommand) IsAutoGroupType() {}

func ProvideQueryCommand(commands []QueryCommand) RootCommand {
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		rpc.ValidatorCommand(),
		rpc.BlockCommand(),
	)

	for _, c := range commands {
		if c.Command != nil {
			cmd.AddCommand(c.Command)
		}
	}

	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return RootCommand{cmd}
}
