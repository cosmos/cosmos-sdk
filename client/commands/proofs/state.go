package proofs

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

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
	prove := !viper.GetBool(commands.FlagTrustNode)

	val, h, err := Get(key, prove)
	if err != nil {
		return err
	}
	return OutputProof(val, h)
}
