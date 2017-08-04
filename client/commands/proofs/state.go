package proofs

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	wire "github.com/tendermint/go-wire"
	"github.com/tendermint/go-wire/data"
	lc "github.com/tendermint/light-client"
	"github.com/tendermint/merkleeyes/iavl"

	"github.com/tendermint/basecoin/client/commands"
)

// KeyQueryCmd - CLI command to query a state by key with proof
var KeyQueryCmd = &cobra.Command{
	Use:   "key [key]",
	Short: "Handle proofs for state of abci app",
	Long: `This will look up a given key in the abci app, verify the proof,
and output it as hex.

If you want json output, use an app-specific command that knows key and value structure.`,
	RunE: commands.RequireInit(keyQueryCmd),
}

// Note: we cannot yse GetAndParseAppProof here, as we don't use go-wire to
// parse the object, but rather return the raw bytes
func keyQueryCmd(cmd *cobra.Command, args []string) error {
	// parse cli
	key, err := ParseHexKey(args, "key")
	if err != nil {
		return err
	}
	prove := viper.GetBool(commands.FlagTrustNode)

	// get the proof -> this will be used by all prover commands
	node := commands.GetNode()

	////////////////
	resp, err := node.ABCIQuery("/key", key, prove)
	if err != nil {
		return err
	}
	ph := int(resp.Height)

	// short-circuit with no proofs
	if !prove {
		return OutputProof(data.Bytes(resp.Value), resp.Height)
	}

	// make sure the proof is the proper height
	if !resp.Code.IsOK() {
		return errors.Errorf("Query error %d: %s", resp.Code, resp.Code.String())
	}
	// TODO: Handle null proofs
	if len(resp.Key) == 0 || len(resp.Value) == 0 || len(resp.Proof) == 0 {
		return lc.ErrNoData()
	}
	if ph != 0 && ph != int(resp.Height) {
		return lc.ErrHeightMismatch(ph, int(resp.Height))
	}

	check, err := GetCertifiedCheckpoint(ph)
	if err != nil {
		return err
	}

	proof := new(iavl.KeyExistsProof)
	err = wire.ReadBinaryBytes(resp.Proof, &proof)
	if err != nil {
		return err
	}

	// validate the proof against the certified header to ensure data integrity
	err = proof.Verify(resp.Key, resp.Value, check.Header.AppHash)
	if err != nil {
		return err
	}

	// state just returns raw hex....
	info := data.Bytes(resp.Value)

	// we can reuse this output for other commands for text/json
	// unless they do something special like store a file to disk
	return OutputProof(info, resp.Height)
}
