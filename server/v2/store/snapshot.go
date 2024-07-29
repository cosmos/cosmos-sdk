package store

import (
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"cosmossdk.io/log"
	serverv2 "cosmossdk.io/server/v2"
	"cosmossdk.io/store/v2/snapshots"
	storev2 "cosmossdk.io/store/v2"
)

// QueryBlockResultsCmd implements the default command for a BlockResults query.
func (s *StoreComponent[T]) ExportSnapshotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export app state to snapshot store",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := serverv2.GetViperFromCmd(cmd)

			height, err := cmd.Flags().GetInt64("height")
			if err != nil {
				return err
			}

			logger := log.NewLogger(cmd.OutOrStdout())
			// app := appCreator(logger, db, nil, viper)
			rootStore, _, err := createRootStore(cmd, v, logger)
			if err != nil {
				return err
			}
			if height == 0 {
				lastCommitId, err := rootStore.LastCommitID()
				if err != nil {
					return err
				}
				height = int64(lastCommitId.Version)
			}

			cmd.Printf("Exporting snapshot for height %d\n", height)

			sm, err := createSnapshotsManager(cmd, v, logger, rootStore)
			if err != nil {
				return err
			}

			snapshot, err := sm.Create(uint64(height))
			if err != nil {
				return err
			}

			cmd.Printf("Snapshot created at height %d, format %d, chunks %d\n", snapshot.Height, snapshot.Format, snapshot.Chunks)
			return nil
		},
	}

	addSnapshotFlagsToCmd(cmd)
	cmd.Flags().Int64("height", 0, "Height to export, default to latest state height")

	return cmd
}

// RestoreSnapshotCmd returns a command to restore a snapshot
func (s *StoreComponent[T]) RestoreSnapshotCmd(newApp serverv2.AppCreator[T]) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore <height> <format>",
		Short: "Restore app state from local snapshot",
		Long:  "Restore app state from local snapshot",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := serverv2.GetViperFromCmd(cmd)

			height, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}
			format, err := strconv.ParseUint(args[1], 10, 32)
			if err != nil {
				return err
			}

			logger := log.NewLogger(cmd.OutOrStdout())
			app := newApp(logger, v)
			rootStore := app.GetStore().(storev2.RootStore)
			
			sm, err := createSnapshotsManager(cmd, v, logger, rootStore)
			if err != nil {
				return err
			}

			return sm.RestoreLocalSnapshot(height, uint32(format))
		},
	}

	addSnapshotFlagsToCmd(cmd)

	return cmd
}

func createSnapshotsManager(cmd *cobra.Command, v *viper.Viper, logger log.Logger, store storev2.RootStore) (*snapshots.Manager, error) {
	home := v.GetString(serverv2.FlagHome)
	snapshotStore, err := snapshots.NewStore(filepath.Join(home, "data", "snapshots"))
	if err != nil {
		return nil, err
	}
	var interval, keepRecent uint64
	// if flag was not passed, use as 0.
	if cmd.Flags().Changed(FlagKeepRecent) {
		keepRecent, err = cmd.Flags().GetUint64(FlagKeepRecent)
		if err != nil {
			return nil, err
		}
	}
	if cmd.Flags().Changed(FlagInterval) {
		interval, err = cmd.Flags().GetUint64(FlagInterval)
		if err != nil {
			return nil, err
		}
	}

	sm := snapshots.NewManager(snapshotStore, snapshots.NewSnapshotOptions(interval, uint32(keepRecent)), store.GetStateCommitment().(snapshots.CommitSnapshotter), store.GetStateStorage().(snapshots.StorageSnapshotter), nil, logger)
	return sm, nil
}

func addSnapshotFlagsToCmd(cmd *cobra.Command) {
	cmd.Flags().Uint64(FlagKeepRecent, 0, "KeepRecent defines how many snapshots to keep in heights")
	cmd.Flags().Uint64(FlagInterval, 0, "Interval defines at which heights the snapshot is taken")
}
