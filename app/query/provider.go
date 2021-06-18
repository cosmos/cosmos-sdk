package query

import (
	"github.com/spf13/cobra"
	"go.uber.org/dig"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/rpc"
)

type Inputs struct {
	dig.In

	Commands []*cobra.Command `group:"query"`
}

type Outputs struct {
	dig.Out

	Command *cobra.Command `group:"root"`
}

func Provider(inputs Inputs) Outputs {
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

	for _, c := range inputs.Commands {
		if c != nil {
			cmd.AddCommand(c)
		}
	}

	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return Outputs{
		Command: cmd,
	}
}
