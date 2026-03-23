package main

import (
	"os"

	"cosmossdk.io/log"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/iavlx/internal"
)

func newRollbackCmd() *cobra.Command {
	var backupDir string
	var targetVersion uint64
	cmd := &cobra.Command{
		Use:     "rollback [dir] --version [version]",
		Aliases: []string{"v"},
		Short:   "Interactively browse IAVL store data",
		Args:    cobra.ExactArgs(1),
	}
	cmd.Flags().StringVar(&backupDir, "backup-dir", "", "The directory to store the backup of the current data before rolling back, defaults to [dir]/bak-[timestamp]")
	cmd.Flags().Uint64Var(&targetVersion, "version", 0, "The target version to roll back to")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return internal.RollbackMultiTree(args[0], targetVersion, log.NewLogger(os.Stdout), backupDir)
	}
	return cmd
}
