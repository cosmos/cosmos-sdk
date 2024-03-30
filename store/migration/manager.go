package migration

import (
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/commitment"
	"cosmossdk.io/store/v2/internal/encoding"
	"cosmossdk.io/store/v2/snapshots"
	"cosmossdk.io/store/v2/storage"
)

const (
	// defaultChannelBufferSize is the default buffer size for the migration stream.
	defaultChannelBufferSize = 1024
	// defaultStorageBufferSize is the default buffer size for the storage snapshotter.
	defaultStorageBufferSize = 1024

	migrateChangesetKeyFmt = "m/cs_%x" // m/cs_<version>
)

// VersionedChangeset is a pair of version and Changeset.
type VersionedChangeset struct {
	Version   uint64
	Changeset *corestore.Changeset
}

// Manager manages the migration of the whole state from store/v1 to store/v2.
type Manager struct {
	logger           log.Logger
	snapshotsManager *snapshots.Manager

	stateStorage    *storage.StorageStore
	stateCommitment *commitment.CommitStore

	db              store.RawDB
	mtx             sync.Mutex // mutex for migratedVersion
	migratedVersion uint64

	chChangeset <-chan *VersionedChangeset
	chDone      <-chan struct{}
}

// NewManager returns a new Manager.
func NewManager(db store.RawDB, sm *snapshots.Manager, ss *storage.StorageStore, sc *commitment.CommitStore, logger log.Logger) *Manager {
	return &Manager{
		logger:           logger,
		snapshotsManager: sm,
		stateStorage:     ss,
		stateCommitment:  sc,
		db:               db,
	}
}

// Start starts the whole migration process.
// It migrates the whole state at the given version to the new store/v2 (both SC and SS).
// It also catches up the Changesets which are committed while the migration is in progress.
// `chChangeset` is the channel to receive the committed Changesets from the RootStore.
// `chDone` is the channel to receive the done signal from the RootStore.
// NOTE: It should be called by the RootStore, running in the background.
func (m *Manager) Start(version uint64, chChangeset <-chan *VersionedChangeset, chDone <-chan struct{}) error {
	m.chChangeset = chChangeset
	m.chDone = chDone

	go func() {
		if err := m.writeChangeset(); err != nil {
			m.logger.Error("failed to write changeset", "err", err)
		}
	}()

	if err := m.Migrate(version); err != nil {
		return fmt.Errorf("failed to migrate state: %w", err)
	}

	return m.Sync()
}

// GetStateCommitment returns the state commitment.
func (m *Manager) GetStateCommitment() *commitment.CommitStore {
	return m.stateCommitment
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
	chStorage := make(chan *corestore.StateChanges, defaultStorageBufferSize)

	eg := new(errgroup.Group)
	eg.Go(func() error {
		return m.stateStorage.Restore(height, chStorage)
	})
	eg.Go(func() error {
		defer close(chStorage)
		_, err := m.stateCommitment.Restore(height, 0, ms, chStorage)
		return err
	})

	if err := eg.Wait(); err != nil {
		return err
	}

	m.mtx.Lock()
	m.migratedVersion = height
	m.mtx.Unlock()

	return nil
}

// writeChangeset writes the Changeset to the db.
func (m *Manager) writeChangeset() error {
	for vc := range m.chChangeset {
		cs := vc.Changeset
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, vc.Version)
		csKey := []byte(fmt.Sprintf(migrateChangesetKeyFmt, buf))
		csBytes, err := encoding.MarshalChangeset(cs)
		if err != nil {
			return fmt.Errorf("failed to marshal changeset: %w", err)
		}

		batch := m.db.NewBatch()
		defer batch.Close()

		if err := batch.Set(csKey, csBytes); err != nil {
			return fmt.Errorf("failed to write changeset to db.Batch: %w", err)
		}
		if err := batch.Write(); err != nil {
			return fmt.Errorf("failed to write changeset to db: %w", err)
		}
	}

	return nil
}

// GetMigratedVersion returns the migrated version.
// It is used to check the migrated version in the RootStore.
func (m *Manager) GetMigratedVersion() uint64 {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	return m.migratedVersion
}

// Sync catches up the Changesets which are committed while the migration is in progress.
// It should be called after the migration is done.
func (m *Manager) Sync() error {
	version := m.GetMigratedVersion()
	if version == 0 {
		return fmt.Errorf("migration is not done yet")
	}
	version += 1

	for {
		select {
		case <-m.chDone:
			return nil
		default:
			buf := make([]byte, 8)
			binary.BigEndian.PutUint64(buf, version)
			csKey := []byte(fmt.Sprintf(migrateChangesetKeyFmt, buf))
			csBytes, err := m.db.Get(csKey)
			if err != nil {
				return fmt.Errorf("failed to get changeset from db: %w", err)
			}
			if csBytes == nil {
				// wait for the next changeset
				time.Sleep(100 * time.Millisecond)
				continue
			}

			cs := corestore.NewChangeset()
			if err := encoding.UnmarshalChangeset(cs, csBytes); err != nil {
				return fmt.Errorf("failed to unmarshal changeset: %w", err)
			}

			if err := m.stateCommitment.WriteBatch(cs); err != nil {
				return fmt.Errorf("failed to write changeset to commitment: %w", err)
			}
			if _, err := m.stateCommitment.Commit(version); err != nil {
				return fmt.Errorf("failed to commit changeset to commitment: %w", err)
			}
			if err := m.stateStorage.ApplyChangeset(version, cs); err != nil {
				return fmt.Errorf("failed to write changeset to storage: %w", err)
			}

			m.mtx.Lock()
			m.migratedVersion = version
			m.mtx.Unlock()

			version += 1
		}
	}
}

// Close closes the manager. It should be called after the migration is done.
// It will close the db and notify the snapshotsManager that the migration is done.
func (m *Manager) Close() error {
	if err := m.db.Close(); err != nil {
		return fmt.Errorf("failed to close db: %w", err)
	}
	m.snapshotsManager.EndMigration(m.stateCommitment)

	return nil
}
