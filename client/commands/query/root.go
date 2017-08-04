package query

import (
	"github.com/spf13/cobra"
	"github.com/tendermint/basecoin/client/commands"
)

// nolint
const (
	FlagHeight = "height"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "query",
	Short: "Get and store merkle proofs for blockchain data",
	Long: `Proofs allows you to validate data and merkle proofs.

These proofs tie the data to a checkpoint, which is managed by "seeds".
Here we can validate these proofs and import/export them to prove specific
data to other peers as needed.
`,
}

func init() {
	RootCmd.PersistentFlags().Int(FlagHeight, 0, "Height to query (skip to use latest block)")
	RootCmd.PersistentFlags().Bool(commands.FlagTrustNode, false,
		"DANGEROUS: blindly trust all results from the server")
	RootCmd.PersistentFlags().MarkHidden(commands.FlagTrustNode)
}
