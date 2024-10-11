package store

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"cosmossdk.io/log"
	serverv2 "cosmossdk.io/server/v2"
	storev2 "cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/snapshots"
	"cosmossdk.io/store/v2/snapshots/types"
)

const SnapshotFileName = "_snapshot"

// ExportSnapshotCmd exports app state to snapshot store.
func (s *Server[T]) ExportSnapshotCmd() *cobra.Command {
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
			rootStore, _, err := createRootStore(v, logger)
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
func (s *Server[T]) RestoreSnapshotCmd(rootStore storev2.Backend) *cobra.Command {
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

// ListSnapshotsCmd returns the command to list local snapshots
func (s *Server[T]) ListSnapshotsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List local snapshots",
		RunE: func(cmd *cobra.Command, args []string) error {
			v := serverv2.GetViperFromCmd(cmd)
			snapshotStore, err := snapshots.NewStore(filepath.Join(v.GetString(serverv2.FlagHome), "data", "snapshots"))
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

	return cmd
}

// DeleteSnapshotCmd returns the command to delete a local snapshot
func (s *Server[T]) DeleteSnapshotCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <height> <format>",
		Short: "Delete a local snapshot",
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

			snapshotStore, err := snapshots.NewStore(filepath.Join(v.GetString(serverv2.FlagHome), "data", "snapshots"))
			if err != nil {
				return err
			}

			return snapshotStore.Delete(height, uint32(format))
		},
	}
}

// DumpArchiveCmd returns a command to dump the snapshot as portable archive format
func (s *Server[T]) DumpArchiveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dump <height> <format>",
		Short: "Dump the snapshot as portable archive format",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := serverv2.GetViperFromCmd(cmd)
			snapshotStore, err := snapshots.NewStore(filepath.Join(v.GetString(serverv2.FlagHome), "data", "snapshots"))
			if err != nil {
				return err
			}

			output, err := cmd.Flags().GetString("output")
			if err != nil {
				return err
			}

			height, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}
			format, err := strconv.ParseUint(args[1], 10, 32)
			if err != nil {
				return err
			}

			if output == "" {
				output = fmt.Sprintf("%d-%d.tar.gz", height, format)
			}

			snapshot, err := snapshotStore.Get(height, uint32(format))
			if err != nil {
				return err
			}

			if snapshot == nil {
				return errors.New("snapshot doesn't exist")
			}

			bz, err := snapshot.Marshal()
			if err != nil {
				return err
			}

			fp, err := os.Create(output)
			if err != nil {
				return err
			}
			defer fp.Close()

			// since the chunk files are already compressed, we just use fastest compression here
			gzipWriter, err := gzip.NewWriterLevel(fp, gzip.BestSpeed)
			if err != nil {
				return err
			}
			tarWriter := tar.NewWriter(gzipWriter)
			if err := tarWriter.WriteHeader(&tar.Header{
				Name: SnapshotFileName,
				Mode: 0o644,
				Size: int64(len(bz)),
			}); err != nil {
				return fmt.Errorf("failed to write snapshot header to tar: %w", err)
			}
			if _, err := tarWriter.Write(bz); err != nil {
				return fmt.Errorf("failed to write snapshot to tar: %w", err)
			}

			for i := uint32(0); i < snapshot.Chunks; i++ {
				path := snapshotStore.PathChunk(height, uint32(format), i)
				tarName := strconv.FormatUint(uint64(i), 10)
				if err := processChunk(tarWriter, path, tarName); err != nil {
					return err
				}
			}

			if err := tarWriter.Close(); err != nil {
				return fmt.Errorf("failed to close tar writer: %w", err)
			}

			if err := gzipWriter.Close(); err != nil {
				return fmt.Errorf("failed to close gzip writer: %w", err)
			}

			return fp.Close()
		},
	}

	cmd.Flags().StringP("output", "o", "", "output file")

	return cmd
}

// LoadArchiveCmd load a portable archive format snapshot into snapshot store
func (s *Server[T]) LoadArchiveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "load <archive-file>",
		Short: "Load a snapshot archive file (.tar.gz) into snapshot store",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v := serverv2.GetViperFromCmd(cmd)
			snapshotStore, err := snapshots.NewStore(filepath.Join(v.GetString(serverv2.FlagHome), "data", "snapshots"))
			if err != nil {
				return err
			}

			path := args[0]
			fp, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open archive file: %w", err)
			}
			reader, err := gzip.NewReader(fp)
			if err != nil {
				return fmt.Errorf("failed to create gzip reader: %w", err)
			}

			var snapshot types.Snapshot
			tr := tar.NewReader(reader)

			hdr, err := tr.Next()
			if err != nil {
				return fmt.Errorf("failed to read snapshot file header: %w", err)
			}
			if hdr.Name != SnapshotFileName {
				return fmt.Errorf("invalid archive, expect file: snapshot, got: %s", hdr.Name)
			}
			bz, err := io.ReadAll(tr)
			if err != nil {
				return fmt.Errorf("failed to read snapshot file: %w", err)
			}
			if err := snapshot.Unmarshal(bz); err != nil {
				return fmt.Errorf("failed to unmarshal snapshot: %w", err)
			}

			// make sure the channel is unbuffered, because the tar reader can't do concurrency
			chunks := make(chan io.ReadCloser)
			quitChan := make(chan *types.Snapshot)
			go func() {
				defer close(quitChan)

				savedSnapshot, err := snapshotStore.Save(snapshot.Height, snapshot.Format, chunks)
				if err != nil {
					cmd.Println("failed to save snapshot", err)
					return
				}
				quitChan <- savedSnapshot
			}()

			for i := uint32(0); i < snapshot.Chunks; i++ {
				hdr, err = tr.Next()
				if err != nil {
					if errors.Is(err, io.EOF) {
						break
					}
					return err
				}

				if hdr.Name != strconv.FormatInt(int64(i), 10) {
					return fmt.Errorf("invalid archive, expect file: %d, got: %s", i, hdr.Name)
				}

				bz, err := io.ReadAll(tr)
				if err != nil {
					return fmt.Errorf("failed to read chunk file: %w", err)
				}
				chunks <- io.NopCloser(bytes.NewReader(bz))
			}
			close(chunks)

			savedSnapshot := <-quitChan
			if savedSnapshot == nil {
				return errors.New("failed to save snapshot")
			}

			if !reflect.DeepEqual(&snapshot, savedSnapshot) {
				_ = snapshotStore.Delete(snapshot.Height, snapshot.Format)
				return errors.New("invalid archive, the saved snapshot is not equal to the original one")
			}

			return nil
		},
	}
}

func createSnapshotsManager(
	cmd *cobra.Command, v *viper.Viper, logger log.Logger, store storev2.Backend,
) (*snapshots.Manager, error) {
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

	sm := snapshots.NewManager(
		snapshotStore, snapshots.NewSnapshotOptions(interval, uint32(keepRecent)),
		store.GetStateCommitment().(snapshots.CommitSnapshotter),
		store.GetStateStorage().(snapshots.StorageSnapshotter),
		nil, logger)
	return sm, nil
}

func addSnapshotFlagsToCmd(cmd *cobra.Command) {
	cmd.Flags().Uint64(FlagKeepRecent, 0, "KeepRecent defines how many snapshots to keep in heights")
	cmd.Flags().Uint64(FlagInterval, 0, "Interval defines at which heights the snapshot is taken")
}

func processChunk(tarWriter *tar.Writer, path, tarName string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open chunk file %s: %w", path, err)
	}
	defer file.Close()

	st, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat chunk file %s: %w", path, err)
	}

	if err := tarWriter.WriteHeader(&tar.Header{
		Name: tarName,
		Mode: 0o644,
		Size: st.Size(),
	}); err != nil {
		return fmt.Errorf("failed to write chunk header to tar: %w", err)
	}

	if _, err := io.Copy(tarWriter, file); err != nil {
		return fmt.Errorf("failed to write chunk to tar: %w", err)
	}

	return nil
}
