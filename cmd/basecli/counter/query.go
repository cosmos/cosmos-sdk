package counter

import (
	"github.com/spf13/cobra"

	wire "github.com/tendermint/go-wire"
	"github.com/tendermint/light-client/commands"
	proofcmd "github.com/tendermint/light-client/commands/proofs"
	"github.com/tendermint/light-client/proofs"

	"github.com/tendermint/basecoin/plugins/counter"
)

var CounterQueryCmd = &cobra.Command{
	Use:   "counter",
	Short: "Query counter state, with proof",
	RunE:  doCounterQuery,
}

func doCounterQuery(cmd *cobra.Command, args []string) error {
	height := proofcmd.GetHeight()
	key := counter.New().StateKey()

	node := commands.GetNode()
	prover := proofs.NewAppProver(node)
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

/*** doesn't seem to be needed anymore??? ***/

// type CounterPresenter struct{}

// func (_ CounterPresenter) MakeKey(str string) ([]byte, error) {
//   key := counter.New().StateKey()
//   return key, nil
// }

// func (_ CounterPresenter) ParseData(raw []byte) (interface{}, error) {
//   var cp counter.CounterPluginState
//   err := wire.ReadBinaryBytes(raw, &cp)
//   return cp, err
// }
