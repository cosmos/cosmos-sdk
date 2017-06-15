package counter

import (
	"github.com/spf13/cobra"

	proofcmd "github.com/tendermint/light-client/commands/proofs"

	"github.com/tendermint/basecoin/plugins/counter"
)

var CounterQueryCmd = &cobra.Command{
	Use:   "counter",
	Short: "Query counter state, with proof",
	RunE:  doCounterQuery,
}

func doCounterQuery(cmd *cobra.Command, args []string) error {
	key := counter.New().StateKey()

	var cp counter.CounterPluginState
	proof, err := proofcmd.GetAndParseAppProof(key, &cp)
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
