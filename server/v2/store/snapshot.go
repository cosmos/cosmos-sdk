package store

import (
	"path/filepath"

	"github.com/spf13/cobra"

	"cosmossdk.io/log"
	serverv2 "cosmossdk.io/server/v2"
	"cosmossdk.io/store/v2/snapshots"
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

			home := v.GetString(serverv2.FlagHome)

			logger := log.NewLogger(cmd.OutOrStdout())
			// app := appCreator(logger, db, nil, viper)
			rootStore, _, err := createRootStore(cmd, home, v, logger)
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

			snapshotStore, err := snapshots.NewStore(filepath.Join(home, "data", "snapshots"))
			if err != nil {
				return err
			}

			var interval, keepRecent uint64
			// if flag was not passed, use as 0.
			if cmd.Flags().Changed(FlagKeepRecent) {
				keepRecent, err = cmd.Flags().GetUint64(FlagKeepRecent)
				if err != nil {
					return err
				}
			}
			if cmd.Flags().Changed(FlagInterval) {
				interval, err = cmd.Flags().GetUint64(FlagInterval)
				if err != nil {
					return err
				}
			}

			sm := snapshots.NewManager(snapshotStore, snapshots.NewSnapshotOptions(interval, uint32(keepRecent)), rootStore.GetStateCommitment().(snapshots.CommitSnapshotter), rootStore.GetStateStorage().(snapshots.StorageSnapshotter), nil, logger)
			snapshot, err := sm.Create(uint64(height))
			if err != nil {
				return err
			}

			cmd.Printf("Snapshot created at height %d, format %d, chunks %d\n", snapshot.Height, snapshot.Format, snapshot.Chunks)
			return nil
		},
	}

	cmd.Flags().Int64("height", 0, "Height to export, default to latest state height")
	cmd.Flags().Uint64(FlagKeepRecent, 0, "KeepRecent defines how many snapshots to keep in heights")
	cmd.Flags().Uint64(FlagInterval, 0, "Interval defines at which heights the snapshot is taken")

	return cmd
}
