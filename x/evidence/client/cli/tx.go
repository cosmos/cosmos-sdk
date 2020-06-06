package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/x/evidence/types"

	"github.com/spf13/cobra"
)

// GetTxCmd returns a CLI command that has all the native evidence module tx
// commands mounted. In addition, it mounts all childCmds, implemented by outside
// modules, under a sub-command. This allows external modules to implement custom
// Evidence types and Handlers while having the ability to create and sign txs
// containing them all from a single root command.
func GetTxCmd(clientCtx client.Context, childCmds []*cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Evidence transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	submitEvidenceCmd := SubmitEvidenceCmd(clientCtx)
	for _, childCmd := range childCmds {
		submitEvidenceCmd.AddCommand(flags.PostCommands(childCmd)[0])
	}

	// TODO: Add tx commands.

	return cmd
}

// SubmitEvidenceCmd returns the top-level evidence submission command handler.
// All concrete evidence submission child command handlers should be registered
// under this command.
func SubmitEvidenceCmd(_ client.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit",
		Short: "Submit arbitrary evidence of misbehavior",
	}

	return cmd
}
