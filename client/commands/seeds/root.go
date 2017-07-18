package seeds

import "github.com/spf13/cobra"

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "seeds",
	Short: "Verify seeds from your local store",
	Long: `Seeds allows you to inspect and update the validator set for the chain.

Since all security in a PoS system is based on having the correct validator
set, it is important to inspect the seeds to maintain the security, which
is used to verify all header and merkle proofs.
`,
}
