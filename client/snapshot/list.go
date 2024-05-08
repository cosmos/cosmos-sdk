package snapshot

import (
	"fmt"

	"github.com/spf13/cobra"

	corectx "cosmossdk.io/core/context"
	"github.com/cosmos/cosmos-sdk/server"
)

// ListSnapshotsCmd returns the command to list local snapshots
var ListSnapshotsCmd = &cobra.Command{
	Use:   "list",
	Short: "List local snapshots",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := corectx.GetServerContextFromCmd(cmd)
		snapshotStore, err := server.GetSnapshotStore(ctx.GetViper())
		if err != nil {
			return err
		}
		snapshots, err := snapshotStore.List()
		if err != nil {
			return fmt.Errorf("failed to list snapshots: %w", err)
		}
		for _, snapshot := range snapshots {
			cmd.Println("height:", snapshot.Height, "format:", snapshot.Format, "chunks:", snapshot.Chunks)
		}

		return nil
	},
}
