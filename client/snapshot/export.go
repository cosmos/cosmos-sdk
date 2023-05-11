package snapshot

import (
	"fmt"
	"strconv"

	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/spf13/cobra"
)

// ExportSnapshotCmd returns a command to take a snapshot of the application state
func ExportSnapshotCmd(appCreator servertypes.AppCreator) *cobra.Command {
	return &cobra.Command{
		Use:   "export <height>",
		Short: "Export app state to snapshot store",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := server.GetServerContextFromCmd(cmd)

			height, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			home := ctx.Config.RootDir
			db, err := openDB(home, server.GetAppDBBackend(ctx.Viper))
			if err != nil {
				return err
			}
			logger := log.NewLogger(cmd.OutOrStdout())
			app := appCreator(logger, db, nil, ctx.Viper)

			if height == 0 {
				height = uint64(app.CommitMultiStore().LastCommitID().Version)
			}

			sm := app.SnapshotManager()
			snapshot, err := sm.Create(height)
			if err != nil {
				return err
			}

			fmt.Printf("Snapshot created at height %d, format %d, chunks %d\n", snapshot.Height, snapshot.Format, snapshot.Chunks)
			return nil
		},
	}
}
