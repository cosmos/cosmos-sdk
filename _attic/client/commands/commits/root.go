package commits

import "github.com/spf13/cobra"

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "commits",
	Short: "Verify commits from your local store",
	Long: `Commits allows you to inspect and update the validator set for the chain.

Since all security in a PoS system is based on having the correct validator
set, it is important to inspect the commits to maintain the security, which
is used to verify all header and merkle proofs.
`,
}
