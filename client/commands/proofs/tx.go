package proofs

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/light-client/proofs"
	"github.com/tendermint/tendermint/types"

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

// type ResultTx struct {
//   Height   int           `json:"height"`
//   Index    int           `json:"index"`
//   TxResult abci.Result   `json:"tx_result"`
//   Tx       types.Tx      `json:"tx"`
//   Proof    types.TxProof `json:"proof,omitempty"`
// }

// type TxProof struct {
//   Index, Total int
//   RootHash     data.Bytes
//   Data         Tx
//   Proof        merkle.SimpleProof
// }

func txQueryCmd(cmd *cobra.Command, args []string) error {
	// parse cli
	// TODO: when querying historical heights is allowed... pass it
	// height := GetHeight()
	bkey, err := ParseHexKey(args, "txhash")
	if err != nil {
		return err
	}

	// get the proof -> this will be used by all prover commands
	node := commands.GetNode()
	prove := !viper.GetBool(commands.FlagTrustNode)
	res, err := node.Tx(bkey, prove)
	if err != nil {
		return err
	}

	if !prove {
		return showTx(res.Height, res.Tx)
	}

	check, err := GetCertifiedCheckpoint(res.Height)
	if err != nil {
		return err
	}
	err = res.Proof.Validate(check.Header.DataHash)
	if err != nil {
		return err
	}

	return showTx(res.Height, res.Proof.Data)
}

func showTx(h int, tx types.Tx) error {
	// auto-determine which tx it was, over all registered tx types
	info, err := TxPresenters.BruteForce(tx)
	if err != nil {
		return err
	}

	// we can reuse this output for other commands for text/json
	// unless they do something special like store a file to disk
	return OutputProof(info, uint64(h))
}
