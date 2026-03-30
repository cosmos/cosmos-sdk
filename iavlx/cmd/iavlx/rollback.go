package main

import (
	"os"

	"cosmossdk.io/log/v2"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/iavlx/internal"
)

// newRollbackCmd creates the offline rollback command.
// This operates directly on the filesystem while the node is stopped — it truncates WALs,
// rolls back checkpoints, and removes commit info files beyond the target version.
// Original files are moved to a backup directory (not deleted) so the rollback can be undone.
func newRollbackCmd() *cobra.Command {
	var backupDir string
	var targetVersion uint64
	cmd := &cobra.Command{
		Use:     "rollback [dir] --version [version]",
		Short:   "Roll back an iavlx multi-tree to a specific version (offline only — the node must be stopped)",
		Args:    cobra.ExactArgs(1),
	}
	cmd.Flags().StringVar(&backupDir, "backup-dir", "", "The directory to store the backup of the current data before rolling back, defaults to [dir]/bak-[timestamp]")
	cmd.Flags().Uint64Var(&targetVersion, "version", 0, "The target version to roll back to")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return internal.RollbackMultiTree(args[0], targetVersion, log.NewLogger(os.Stdout), backupDir)
	}
	return cmd
}
