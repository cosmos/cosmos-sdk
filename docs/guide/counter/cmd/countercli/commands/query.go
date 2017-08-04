package commands

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/basecoin/client/commands"
	proofcmd "github.com/tendermint/basecoin/client/commands/proofs"

	"github.com/tendermint/basecoin/docs/guide/counter/plugins/counter"
	"github.com/tendermint/basecoin/stack"
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
	h, err := proofcmd.GetParsed(key, &cp, prove)
	if err != nil {
		return err
	}

	return proofcmd.OutputProof(cp, h)
}
