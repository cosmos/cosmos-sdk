package root

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	corelog "cosmossdk.io/core/log"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/metrics"
	"cosmossdk.io/store/v2/migration"
	"cosmossdk.io/store/v2/proof"
	"cosmossdk.io/store/v2/pruning"
)

var (
	_ store.RootStore        = (*Store)(nil)
	_ store.UpgradeableStore = (*Store)(nil)
)

// Store defines the SDK's default RootStore implementation. It contains a single
// State Storage (SS) backend and a single State Commitment (SC) backend. The SC
// backend may or may not support multiple store keys and is implementation
// dependent.
type Store struct {
	logger corelog.Logger

	// holds the db instance for closing it
	dbCloser io.Closer

	// stateCommitment reflects the state commitment (SC) backend
	stateCommitment store.Committer

	// lastCommitInfo reflects the last version/hash that has been committed
	lastCommitInfo *proof.CommitInfo

	// telemetry reflects a telemetry agent responsible for emitting metrics (if any)
	telemetry metrics.StoreMetrics

	// pruningManager reflects the pruning manager used to prune state of the SS and SC backends
	pruningManager *pruning.Manager

	// Migration related fields
	// migrationManager reflects the migration manager used to migrate state from v1 to v2
	migrationManager *migration.Manager
	// chChangeset reflects the channel used to send the changeset to the migration manager
	chChangeset chan *migration.VersionedChangeset
	// chDone reflects the channel used to signal the migration manager that the migration
	// is done
	chDone chan struct{}
	// isMigrating reflects whether the store is currently migrating
	isMigrating bool
}

// New creates a new root Store instance.
//
// NOTE: The migration manager is optional and can be nil if no migration is required.
func New(
	dbCloser io.Closer,
	logger corelog.Logger,
	sc store.Committer,
	pm *pruning.Manager,
	mm *migration.Manager,
	m metrics.StoreMetrics,
) (store.RootStore, error) {
	return &Store{
		dbCloser:         dbCloser,
		logger:           logger,
		stateCommitment:  sc,
		pruningManager:   pm,
		migrationManager: mm,
		telemetry:        m,
		isMigrating:      mm != nil,
	}, nil
}

// Close closes the store and resets all internal fields. Note, Close() is NOT
// idempotent and should only be called once.
func (s *Store) Close() (err error) {
	err = errors.Join(err, s.stateCommitment.Close())
	err = errors.Join(err, s.dbCloser.Close())

	s.stateCommitment = nil
	s.lastCommitInfo = nil

	return err
}

func (s *Store) SetMetrics(m metrics.Metrics) {
	s.telemetry = m
}

func (s *Store) SetInitialVersion(v uint64) error {
	return s.stateCommitment.SetInitialVersion(v)
}

// getVersionedReader returns a VersionedReader based on the given version. If the
// version exists in the state storage, it returns the state storage.
// If not, it checks if the state commitment implements the VersionedReader interface
// and the version exists in the state commitment, since the state storage will be
// synced during migration.
func (s *Store) getVersionedReader(version uint64) (store.VersionedReader, error) {
	isExist, err := s.stateCommitment.VersionExists(version)
	if err != nil {
		return nil, err
	}
	if isExist {
		return s.stateCommitment, nil
	}
	return nil, fmt.Errorf("version %d does not exist", version)
}

func (s *Store) StateLatest() (uint64, corestore.ReaderMap, error) {
	v, err := s.GetLatestVersion()
	if err != nil {
		return 0, nil, err
	}
	vReader, err := s.getVersionedReader(v)
	if err != nil {
		return 0, nil, err
	}

	return v, NewReaderMap(v, vReader), nil
}

// StateAt returns a read-only view of the state at a given version.
func (s *Store) StateAt(v uint64) (corestore.ReaderMap, error) {
	vReader, err := s.getVersionedReader(v)
	return NewReaderMap(v, vReader), err
}

func (s *Store) GetStateCommitment() store.Committer {
	return s.stateCommitment
}

// LastCommitID returns a CommitID based off of the latest internal CommitInfo.
// If an internal CommitInfo is not set, a new one will be returned with only the
// latest version set, which is based off of the SC view.
func (s *Store) LastCommitID() (proof.CommitID, error) {
	if s.lastCommitInfo != nil {
		return s.lastCommitInfo.CommitID(), nil
	}

	latestVersion, err := s.stateCommitment.GetLatestVersion()
	if err != nil {
		return proof.CommitID{}, err
	}
	// if the latest version is 0, we return a CommitID with version 0 and a hash of an empty byte slice
	bz := sha256.Sum256([]byte{})

	return proof.CommitID{Version: latestVersion, Hash: bz[:]}, nil
}

// GetLatestVersion returns the latest version based on the latest internal
// CommitInfo. An error is returned if the latest CommitInfo or version cannot
// be retrieved.
func (s *Store) GetLatestVersion() (uint64, error) {
	lastCommitID, err := s.LastCommitID()
	if err != nil {
		return 0, err
	}

	return lastCommitID.Version, nil
}

func (s *Store) Query(storeKey []byte, version uint64, key []byte, prove bool) (store.QueryResult, error) {
	if s.telemetry != nil {
		now := time.Now()
		defer s.telemetry.MeasureSince(now, "root_store", "query")
	}

	val, err := s.stateCommitment.Get(storeKey, version, key)
	if err != nil {
		return store.QueryResult{}, fmt.Errorf("failed to query SC store: %w", err)
	}

	result := store.QueryResult{
		Key:     key,
		Value:   val,
		Version: version,
	}

	if prove {
		result.ProofOps, err = s.stateCommitment.GetProof(storeKey, version, key)
		if err != nil {
			return store.QueryResult{}, fmt.Errorf("failed to get SC store proof: %w", err)
		}
	}

	return result, nil
}

func (s *Store) LoadLatestVersion() error {
	if s.telemetry != nil {
		now := time.Now()
		defer s.telemetry.MeasureSince(now, "root_store", "load_latest_version")
	}

	lv, err := s.GetLatestVersion()
	if err != nil {
		return err
	}

	return s.loadVersion(lv, nil, false)
}

func (s *Store) LoadVersion(version uint64) error {
	if s.telemetry != nil {
		now := time.Now()
		defer s.telemetry.MeasureSince(now, "root_store", "load_version")
	}

	return s.loadVersion(version, nil, false)
}

func (s *Store) LoadVersionForOverwriting(version uint64) error {
	if s.telemetry != nil {
		now := time.Now()
		defer s.telemetry.MeasureSince(now, "root_store", "load_version_for_overwriting")
	}

	return s.loadVersion(version, nil, true)
}

// LoadVersionAndUpgrade implements the UpgradeableStore interface.
//
// NOTE: It cannot be called while the store is migrating.
func (s *Store) LoadVersionAndUpgrade(version uint64, upgrades *corestore.StoreUpgrades) error {
	if upgrades == nil {
		return errors.New("upgrades cannot be nil")
	}

	if s.telemetry != nil {
		defer s.telemetry.MeasureSince(time.Now(), "root_store", "load_version_and_upgrade")
	}

	if s.isMigrating {
		return errors.New("cannot upgrade while migrating")
	}

	if err := s.loadVersion(version, upgrades, true); err != nil {
		return err
	}

	return nil
}

func (s *Store) loadVersion(v uint64, upgrades *corestore.StoreUpgrades, overrideAfter bool) error {
	s.logger.Debug("loading version", "version", v)

	if upgrades == nil {
		if !overrideAfter {
			if err := s.stateCommitment.LoadVersion(v); err != nil {
				return fmt.Errorf("failed to load SC version %d: %w", v, err)
			}
		} else {
			if err := s.stateCommitment.LoadVersionForOverwriting(v); err != nil {
				return fmt.Errorf("failed to load SC version %d: %w", v, err)
			}
		}
	} else {
		// if upgrades are provided, we need to load the version and apply the upgrades
		if err := s.stateCommitment.LoadVersionAndUpgrade(v, upgrades); err != nil {
			return fmt.Errorf("failed to load SS version with upgrades %d: %w", v, err)
		}
	}

	// set lastCommitInfo explicitly s.t. Commit commits the correct version, i.e. v+1
	var err error
	s.lastCommitInfo, err = s.stateCommitment.GetCommitInfo(v)
	if err != nil {
		return fmt.Errorf("failed to get commit info for version %d: %w", v, err)
	}

	// if we're migrating, we need to start the migration process
	if s.isMigrating {
		s.startMigration()
	}

	return nil
}

// Commit commits all state changes to the underlying SS and SC backends. It
// writes a batch of the changeset to the SC tree, and retrieves the CommitInfo
// from the SC tree. Finally, it commits the SC tree and returns the hash of
// the CommitInfo.
func (s *Store) Commit(cs *corestore.Changeset) ([]byte, error) {
	if s.telemetry != nil {
		now := time.Now()
		defer s.telemetry.MeasureSince(now, "root_store", "commit")
	}

	if err := s.handleMigration(cs); err != nil {
		return nil, err
	}

	// signal to the pruning manager that a new version is about to be committed
	// this may be required if the SS and SC backends implementation have the
	// background pruning process (iavl v1 for example) which must be paused during the commit
	s.pruningManager.PausePruning()

	var cInfo *proof.CommitInfo
	if err := s.stateCommitment.WriteChangeset(cs); err != nil {
		return nil, fmt.Errorf("failed to write batch to SC store: %w", err)
	}

	cInfo, err := s.stateCommitment.Commit(cs.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to commit SC store: %w", err)
	}

	if cInfo.Version != cs.Version {
		return nil, fmt.Errorf("commit version mismatch: got %d, expected %d", cInfo.Version, cs.Version)
	}
	s.lastCommitInfo = cInfo

	// signal to the pruning manager that the commit is done
	if err := s.pruningManager.ResumePruning(s.lastCommitInfo.Version); err != nil {
		s.logger.Error("failed to signal commit done to pruning manager", "err", err)
	}

	return s.lastCommitInfo.Hash(), nil
}

// startMigration starts a migration process to migrate the RootStore/v1 to the
// SS and SC backends of store/v2 and initializes the channels.
// It runs in a separate goroutine and replaces the current RootStore with the
// migrated new backends once the migration is complete.
//
// NOTE: This method should only be called once after loadVersion.
func (s *Store) startMigration() {
	// buffer at most 1 changeset, if the receiver is behind attempting to buffer
	// more than 1 will block.
	s.chChangeset = make(chan *migration.VersionedChangeset, 1)
	// it is used to signal the migration manager that the migration is done
	s.chDone = make(chan struct{})

	mtx := sync.Mutex{}
	mtx.Lock()
	go func() {
		version := s.lastCommitInfo.Version
		s.logger.Info("starting migration", "version", version)
		mtx.Unlock()
		if err := s.migrationManager.Start(version, s.chChangeset, s.chDone); err != nil {
			s.logger.Error("failed to start migration", "err", err)
		}
	}()

	// wait for the migration manager to start
	mtx.Lock()
	defer mtx.Unlock()
}

func (s *Store) handleMigration(cs *corestore.Changeset) error {
	if s.isMigrating {
		// if the migration manager has already migrated to the version, close the
		// channels and replace the state commitment
		if s.migrationManager.GetMigratedVersion() == s.lastCommitInfo.Version {
			close(s.chDone)
			close(s.chChangeset)
			s.isMigrating = false
			// close the old state commitment and replace it with the new one
			if err := s.stateCommitment.Close(); err != nil {
				return fmt.Errorf("failed to close the old SC store: %w", err)
			}
			newStateCommitment := s.migrationManager.GetStateCommitment()
			if newStateCommitment != nil {
				s.stateCommitment = newStateCommitment
			}
			if err := s.migrationManager.Close(); err != nil {
				return fmt.Errorf("failed to close migration manager: %w", err)
			}
			s.logger.Info("migration completed", "version", s.lastCommitInfo.Version)
		} else {
			// queue the next changeset to the migration manager
			s.chChangeset <- &migration.VersionedChangeset{Version: s.lastCommitInfo.Version + 1, Changeset: cs}
		}
	}
	return nil
}

func (s *Store) Prune(version uint64) error {
	return s.pruningManager.Prune(version)
}
