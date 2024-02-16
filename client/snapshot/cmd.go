package snapshot

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/server/types"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
)

// Cmd returns the snapshots group command
func Cmd[T types.Application](appCreator servertypes.AppCreator[T]) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshots",
		Short: "Manage local snapshots",
	}
	cmd.AddCommand(
		ListSnapshotsCmd,
		RestoreSnapshotCmd(appCreator),
		ExportSnapshotCmd(appCreator),
		DumpArchiveCmd(),
		LoadArchiveCmd(),
		DeleteSnapshotCmd(),
	)
	return cmd
}
