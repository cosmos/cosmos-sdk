package query

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/basecoin"
	wire "github.com/tendermint/go-wire"
	"github.com/tendermint/tendermint/types"

	"github.com/tendermint/basecoin/client/commands"
)

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

	// no checks if we don't get a proof
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

	// note that we return res.Proof.Data, not res.Tx,
	// as res.Proof.Validate only verifies res.Proof.Data
	return showTx(res.Height, res.Proof.Data)
}

// showTx parses anything that was previously registered as basecoin.Tx
func showTx(h int, tx types.Tx) error {
	var info basecoin.Tx
	err := wire.ReadBinaryBytes(tx, &info)
	if err != nil {
		return err
	}
	return OutputProof(info, uint64(h))
}
