package root

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	coreheader "cosmossdk.io/core/header"
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
	logger         corelog.Logger
	initialVersion uint64

	// stateStorage reflects the state storage backend
	stateStorage store.VersionedDatabase

	// stateCommitment reflects the state commitment (SC) backend
	stateCommitment store.Committer

	// commitHeader reflects the header used when committing state
	// note, this isn't required and only used for query purposes)
	commitHeader *coreheader.Info

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
	logger corelog.Logger,
	ss store.VersionedDatabase,
	sc store.Committer,
	pm *pruning.Manager,
	mm *migration.Manager,
	m metrics.StoreMetrics,
) (store.RootStore, error) {
	return &Store{
		logger:           logger,
		initialVersion:   1,
		stateStorage:     ss,
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
	err = errors.Join(err, s.stateStorage.Close())
	err = errors.Join(err, s.stateCommitment.Close())

	s.stateStorage = nil
	s.stateCommitment = nil
	s.lastCommitInfo = nil
	s.commitHeader = nil

	return err
}

func (s *Store) SetMetrics(m metrics.Metrics) {
	s.telemetry = m
}

func (s *Store) SetInitialVersion(v uint64) error {
	s.initialVersion = v

	return s.stateCommitment.SetInitialVersion(v)
}

func (s *Store) StateLatest() (uint64, corestore.ReaderMap, error) {
	v, err := s.GetLatestVersion()
	if err != nil {
		return 0, nil, err
	}

	return v, NewReaderMap(v, s), nil
}

func (s *Store) StateAt(v uint64) (corestore.ReaderMap, error) {
	// TODO(bez): We may want to avoid relying on the SC metadata here. Instead,
	// we should add a VersionExists() method to the VersionedDatabase interface.
	//
	// Ref: https://github.com/cosmos/cosmos-sdk/issues/19091
	if cInfo, err := s.stateCommitment.GetCommitInfo(v); err != nil || cInfo == nil {
		return nil, fmt.Errorf("failed to get commit info for version %d: %w", v, err)
	}

	return NewReaderMap(v, s), nil
}

func (s *Store) GetStateStorage() store.VersionedDatabase {
	return s.stateStorage
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

	var val []byte
	var err error
	if s.isMigrating { // if we're migrating, we need to query the SC backend
		val, err = s.stateCommitment.Get(storeKey, version, key)
		if err != nil {
			return store.QueryResult{}, fmt.Errorf("failed to query SC store: %w", err)
		}
	} else {
		val, err = s.stateStorage.Get(storeKey, version, key)
		if err != nil {
			return store.QueryResult{}, fmt.Errorf("failed to query SS store: %w", err)
		}
		if val == nil {
			// fallback to querying SC backend if not found in SS backend
			//
			// Note, this should only used during migration, i.e. while SS and IAVL v2
			// are being asynchronously synced.
			bz, scErr := s.stateCommitment.Get(storeKey, version, key)
			if scErr != nil {
				return store.QueryResult{}, fmt.Errorf("failed to query SC store: %w", scErr)
			}
			val = bz
		}
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

	return s.loadVersion(lv, nil)
}

func (s *Store) LoadVersion(version uint64) error {
	if s.telemetry != nil {
		now := time.Now()
		defer s.telemetry.MeasureSince(now, "root_store", "load_version")
	}

	return s.loadVersion(version, nil)
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

	if err := s.loadVersion(version, upgrades); err != nil {
		return err
	}

	// if the state storage implements the UpgradableDatabase interface, prune the
	// deleted store keys
	upgradableDatabase, ok := s.stateStorage.(store.UpgradableDatabase)
	if ok {
		if err := upgradableDatabase.PruneStoreKeys(upgrades.Deleted, version); err != nil {
			return fmt.Errorf("failed to prune store keys %v: %w", upgrades.Deleted, err)
		}
	}

	return nil
}

func (s *Store) loadVersion(v uint64, upgrades *corestore.StoreUpgrades) error {
	s.logger.Debug("loading version", "version", v)

	if upgrades == nil {
		if err := s.stateCommitment.LoadVersion(v); err != nil {
			return fmt.Errorf("failed to load SC version %d: %w", v, err)
		}
	} else {
		// if upgrades are provided, we need to load the version and apply the upgrades
		upgradeableStore, ok := s.stateCommitment.(store.UpgradeableStore)
		if !ok {
			return errors.New("SC store does not support upgrades")
		}
		if err := upgradeableStore.LoadVersionAndUpgrade(v, upgrades); err != nil {
			return fmt.Errorf("failed to load SS version with upgrades %d: %w", v, err)
		}
	}

	s.commitHeader = nil

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

func (s *Store) SetCommitHeader(h *coreheader.Info) {
	s.commitHeader = h
}

// WorkingHash writes the changeset to SC and SS and returns the workingHash
// of the CommitInfo.
func (s *Store) WorkingHash(cs *corestore.Changeset) ([]byte, error) {
	if s.telemetry != nil {
		now := time.Now()
		defer s.telemetry.MeasureSince(now, "root_store", "working_hash")
	}

	// write the changeset to the SC and SS backends
	eg := new(errgroup.Group)
	eg.Go(func() error {
		if err := s.writeSC(cs); err != nil {
			return fmt.Errorf("failed to write SC: %w", err)
		}

		return nil
	})
	eg.Go(func() error {
		if err := s.stateStorage.ApplyChangeset(s.initialVersion, cs); err != nil {
			return fmt.Errorf("failed to commit SS: %w", err)
		}

		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	workingHash := s.lastCommitInfo.Hash()
	s.lastCommitInfo.Version -= 1 // reset lastCommitInfo to allow Commit() to work correctly

	return workingHash, nil
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

	// write the changeset to the SC tree and update lastCommitInfo
	if err := s.writeSC(cs); err != nil {
		return nil, err
	}

	version := s.lastCommitInfo.Version

	if s.commitHeader != nil && uint64(s.commitHeader.Height) != version {
		s.logger.Debug("commit header and version mismatch", "header_height", s.commitHeader.Height, "version", version)
	}

	// signal to the pruning manager that a new version is about to be committed
	// this may be required if the SS and SC backends implementation have the
	// background pruning process which must be paused during the commit
	if err := s.pruningManager.SignalCommit(true, version); err != nil {
		s.logger.Error("failed to signal commit to pruning manager", "err", err)
	}

	eg := new(errgroup.Group)

	// if we're migrating, we don't want to commit to the state storage to avoid
	// parallel writes
	if !s.isMigrating {
		// commit SS async
		eg.Go(func() error {
			if err := s.stateStorage.ApplyChangeset(version, cs); err != nil {
				return fmt.Errorf("failed to commit SS: %w", err)
			}

			return nil
		})
	}

	// commit SC async
	eg.Go(func() error {
		if err := s.commitSC(); err != nil {
			return fmt.Errorf("failed to commit SC: %w", err)
		}

		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	// signal to the pruning manager that the commit is done
	if err := s.pruningManager.SignalCommit(false, version); err != nil {
		s.logger.Error("failed to signal commit done to pruning manager", "err", err)
	}

	if s.commitHeader != nil {
		s.lastCommitInfo.Timestamp = s.commitHeader.Time
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

// writeSC accepts a Changeset and writes that as a batch to the underlying SC
// tree, which allows us to retrieve the working hash of the SC tree. Finally,
// we construct a *CommitInfo and set that as lastCommitInfo. Note, this should
// only be called once per block!
// If migration is in progress, the changeset is sent to the migration manager.
func (s *Store) writeSC(cs *corestore.Changeset) error {
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
			s.chChangeset <- &migration.VersionedChangeset{Version: s.lastCommitInfo.Version + 1, Changeset: cs}
		}
	}

	if err := s.stateCommitment.WriteChangeset(cs); err != nil {
		return fmt.Errorf("failed to write batch to SC store: %w", err)
	}

	var previousHeight, version uint64
	if s.lastCommitInfo.GetVersion() == 0 && s.initialVersion > 1 {
		// This case means that no commit has been made in the store, we
		// start from initialVersion.
		version = s.initialVersion
	} else {
		// This case can means two things:
		//
		// 1. There was already a previous commit in the store, in which case we
		// 		increment the version from there.
		// 2. There was no previous commit, and initial version was not set, in which
		// 		case we start at version 1.
		previousHeight = s.lastCommitInfo.GetVersion()
		version = previousHeight + 1
	}

	s.lastCommitInfo = s.stateCommitment.WorkingCommitInfo(version)

	return nil
}

// commitSC commits the SC store. At this point, a batch of the current changeset
// should have already been written to the SC via writeSC(). This method solely
// commits that batch. An error is returned if commit fails or the hash of the
// committed state does not match the hash of the working state.
func (s *Store) commitSC() error {
	cInfo, err := s.stateCommitment.Commit(s.lastCommitInfo.Version)
	if err != nil {
		return fmt.Errorf("failed to commit SC store: %w", err)
	}

	if !bytes.Equal(cInfo.Hash(), s.lastCommitInfo.Hash()) {
		return fmt.Errorf("unexpected commit hash; got: %X, expected: %X", cInfo.Hash(), s.lastCommitInfo.Hash())
	}

	return nil
}

func (s *Store) Prune(version uint64) error {
	return s.pruningManager.Prune(version)
}
