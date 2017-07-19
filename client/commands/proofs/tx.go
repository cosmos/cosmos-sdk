package proofs

import (
	"github.com/spf13/cobra"

	"github.com/tendermint/light-client/proofs"

	"github.com/tendermint/basecoin/client/commands"
)

//nolint TODO add description
var TxPresenters = proofs.NewPresenters()

// TxQueryCmd - CLI command to query a transaction with proof
var TxQueryCmd = &cobra.Command{
	Use:   "tx [txhash]",
	Short: "Handle proofs of commited txs",
	Long: `Proofs allows you to validate abci state with merkle proofs.

These proofs tie the data to a checkpoint, which is managed by "seeds".
Here we can validate these proofs and import/export them to prove specific
data to other peers as needed.
`,
	RunE: commands.RequireInit(txQueryCmd),
}

func txQueryCmd(cmd *cobra.Command, args []string) error {
	// parse cli
	height := GetHeight()
	bkey, err := ParseHexKey(args, "txhash")
	if err != nil {
		return err
	}

	// get the proof -> this will be used by all prover commands
	node := commands.GetNode()
	prover := proofs.NewTxProver(node)
	proof, err := GetProof(node, prover, bkey, height)
	if err != nil {
		return err
	}

	// auto-determine which tx it was, over all registered tx types
	info, err := TxPresenters.BruteForce(proof.Data())
	if err != nil {
		return err
	}

	// we can reuse this output for other commands for text/json
	// unless they do something special like store a file to disk
	return OutputProof(info, proof.BlockHeight())
}
