package migration

import (
	"golang.org/x/sync/errgroup"

	"cosmossdk.io/log"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/snapshots"
)

const (
	// defaultChannelBufferSize is the default buffer size for the migration stream.
	defaultChannelBufferSize = 1024
	// defaultStorageBufferSize is the default buffer size for the storage snapshotter.
	defaultStorageBufferSize = 1024
)

// Manager manages the migration of the whole state from store/v1 to store/v2.
type Manager struct {
	logger           log.Logger
	snapshotsManager *snapshots.Manager

	storageSnapshotter snapshots.StorageSnapshotter
	commitSnapshotter  snapshots.CommitSnapshotter
}

// NewManager returns a new Manager.
func NewManager(sm *snapshots.Manager, ss snapshots.StorageSnapshotter, cs snapshots.CommitSnapshotter, logger log.Logger) *Manager {
	return &Manager{
		logger:             logger,
		snapshotsManager:   sm,
		storageSnapshotter: ss,
		commitSnapshotter:  cs,
	}
}

// Migrate migrates the whole state at the given height to the new store/v2.
func (m *Manager) Migrate(height uint64) error {
	// create the migration stream and snapshot,
	// which acts as protoio.Reader and snapshots.WriteCloser.
	ms := NewMigrationStream(defaultChannelBufferSize)

	if err := m.snapshotsManager.CreateMigration(height, ms); err != nil {
		return err
	}

	// restore the snapshot
	chStorage := make(chan *store.KVPair, defaultStorageBufferSize)

	eg := new(errgroup.Group)
	eg.Go(func() error {
		return m.storageSnapshotter.Restore(height, chStorage)
	})
	eg.Go(func() error {
		defer close(chStorage)
		_, err := m.commitSnapshotter.Restore(height, 0, ms, chStorage)
		return err
	})

	return eg.Wait()
}
