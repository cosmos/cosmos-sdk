package proofs

import (
	"github.com/spf13/cobra"

	"github.com/tendermint/go-wire/data"

	"github.com/tendermint/basecoin/commands"
	"github.com/tendermint/light-client/proofs"
)

var KeyCmd = &cobra.Command{
	Use:   "key [key]",
	Short: "Handle proofs for state of abci app",
	Long: `This will look up a given key in the abci app, verify the proof,
and output it as hex.

If you want json output, use an app-specific command that knows key and value structure.`,
	RunE: commands.RequireInit(doKeyQuery),
}

// Note: we cannot yse GetAndParseAppProof here, as we don't use go-wire to
// parse the object, but rather return the raw bytes
func doKeyQuery(cmd *cobra.Command, args []string) error {
	// parse cli
	height := GetHeight()
	key, err := ParseHexKey(args, "key")
	if err != nil {
		return err
	}

	// get the proof -> this will be used by all prover commands
	node := commands.GetNode()
	prover := proofs.NewAppProver(node)
	proof, err := GetProof(node, prover, key, height)
	if err != nil {
		return err
	}

	// state just returns raw hex....
	info := data.Bytes(proof.Data())

	// we can reuse this output for other commands for text/json
	// unless they do something special like store a file to disk
	return OutputProof(info, proof.BlockHeight())
}
