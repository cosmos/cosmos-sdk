package counter

import (
	"github.com/spf13/cobra"

	wire "github.com/tendermint/go-wire"
	"github.com/tendermint/light-client/commands"
	proofcmd "github.com/tendermint/light-client/commands/proofs"
	"github.com/tendermint/light-client/proofs"

	"github.com/tendermint/basecoin/plugins/counter"
)

var CounterTxCmd = &cobra.Command{
	Use:   "counter",
	Short: "query counter state",
	RunE:  counterTxCmd,
}

func init() {
	//first modify the full node account query command for the light client
	proofcmd.RootCmd.AddCommand(CounterTxCmd)
}

func counterTxCmd(cmd *cobra.Command, args []string) error {

	// get the proof -> this will be used by all prover commands
	height := proofcmd.GetHeight()
	node := commands.GetNode()
	prover := proofs.NewAppProver(node)
	key := counter.New().StateKey()
	proof, err := proofcmd.GetProof(node, prover, key, height)
	if err != nil {
		return err
	}

	var cp counter.CounterPluginState
	err = wire.ReadBinaryBytes(proof.Data(), &cp)
	if err != nil {
		return err
	}

	return proofcmd.OutputProof(cp, proof.BlockHeight())
}
