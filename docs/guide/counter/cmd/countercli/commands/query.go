package commands

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/commands"
	"github.com/cosmos/cosmos-sdk/client/commands/query"

	"github.com/cosmos/cosmos-sdk/docs/guide/counter/plugins/counter"
	"github.com/cosmos/cosmos-sdk/stack"
)

//CounterQueryCmd - CLI command to query the counter state
var CounterQueryCmd = &cobra.Command{
	Use:   "counter",
	Short: "Query counter state, with proof",
	RunE:  counterQueryCmd,
}

func counterQueryCmd(cmd *cobra.Command, args []string) error {
	var cp counter.State

	prove := !viper.GetBool(commands.FlagTrustNode)
	key := stack.PrefixedKey(counter.NameCounter, counter.StateKey())
	h, err := query.GetParsed(key, &cp, prove)
	if err != nil {
		return err
	}

	return query.OutputProof(cp, h)
}
