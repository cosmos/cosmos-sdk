package snapshot

import (
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/spf13/cobra"
)

// Cmd returns the snapshots group command
func Cmd(appCreator servertypes.AppCreator) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshots",
		Short: "Manage local snapshots",
		Long:  "Manage local snapshots",
	}
	cmd.AddCommand(
		ListSnapshotsCmd(appCreator),
		RestoreSnapshotCmd(appCreator),
	)
	return cmd
}
