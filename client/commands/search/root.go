package search

import (
	"github.com/cosmos/cosmos-sdk/client/commands"
	"github.com/spf13/cobra"
)

// nolint
const (
	FlagMinHeight = "minHeight"
	FlagMaxHeight = "maxHeight"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "search",
	Short: "Search for specific transactions based on tags",
	Long: `Search has subcommands for each specific tx type.

Each one has specific flags it allows you to search on and
these are exposed on those subcommands. Returns all historical
transactions that match those tags.
`,
}

func init() {
	RootCmd.PersistentFlags().Int(FlagMinHeight, 0, "Minimum height to include in search (skip for all history)")
	RootCmd.PersistentFlags().Int(FlagMaxHeight, 0, "Max height to include in search (skip to include latest)")
	RootCmd.PersistentFlags().Bool(commands.FlagTrustNode, false,
		"DANGEROUS: blindly trust all results from the server")
	RootCmd.PersistentFlags().MarkHidden(commands.FlagTrustNode)
}
