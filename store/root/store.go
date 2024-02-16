package root

import (
	"bytes"
	"fmt"
	"slices"
	"time"

	"github.com/cockroachdb/errors"
	"golang.org/x/sync/errgroup"

	coreheader "cosmossdk.io/core/header"
	"cosmossdk.io/log"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/metrics"
	"cosmossdk.io/store/v2/proof"
)

var _ store.RootStore = (*Store)(nil)

// Store defines the SDK's default RootStore implementation. It contains a single
// State Storage (SS) backend and a single State Commitment (SC) backend. The SC
// backend may or may not support multiple store keys and is implementation
// dependent.
type Store struct {
	logger         log.Logger
	initialVersion uint64

	// stateStore reflects the state storage backend
	stateStore store.VersionedDatabase

	// stateCommitment reflects the state commitment (SC) backend
	stateCommitment store.Committer

	// commitHeader reflects the header used when committing state (note, this isn't required and only used for query purposes)
	commitHeader *coreheader.Info

	// lastCommitInfo reflects the last version/hash that has been committed
	lastCommitInfo *proof.CommitInfo

	// workingHash defines the current (yet to be committed) hash
	workingHash []byte

	// telemetry reflects a telemetry agent responsible for emitting metrics (if any)
	telemetry metrics.StoreMetrics
}

func New(
	logger log.Logger,
	ss store.VersionedDatabase,
	sc store.Committer,
	m metrics.StoreMetrics,
) (store.RootStore, error) {
	return &Store{
		logger:          logger.With("module", "root_store"),
		initialVersion:  1,
		stateStore:      ss,
		stateCommitment: sc,
		telemetry:       m,
	}, nil
}

// Close closes the store and resets all internal fields. Note, Close() is NOT
// idempotent and should only be called once.
func (s *Store) Close() (err error) {
	err = errors.Join(err, s.stateStore.Close())
	err = errors.Join(err, s.stateCommitment.Close())

	s.stateStore = nil
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

func (s *Store) StateLatest() (uint64, store.ReadOnlyRootStore, error) {
	v, err := s.GetLatestVersion()
	if err != nil {
		return 0, nil, err
	}

	return v, NewReadOnlyAdapter(v, s), nil
}

func (s *Store) StateAt(v uint64) (store.ReadOnlyRootStore, error) {
	// TODO(bez): We may want to avoid relying on the SC metadata here. Instead,
	// we should add a VersionExists() method to the VersionedDatabase interface.
	//
	// Ref: https://github.com/cosmos/cosmos-sdk/issues/19091
	if cInfo, err := s.stateCommitment.GetCommitInfo(v); err != nil || cInfo == nil {
		return nil, fmt.Errorf("failed to get commit info for version %d: %w", v, err)
	}

	return NewReadOnlyAdapter(v, s), nil
}

func (s *Store) GetStateStorage() store.VersionedDatabase {
	return s.stateStore
}

func (s *Store) GetStateCommitment() store.Committer {
	return s.stateCommitment
}

// LastCommitID returns a CommitID based off of the latest internal CommitInfo.
// If an internal CommitInfo is not set, a new one will be returned with only the
// latest version set, which is based off of the SS view.
func (s *Store) LastCommitID() (proof.CommitID, error) {
	if s.lastCommitInfo != nil {
		return s.lastCommitInfo.CommitID(), nil
	}

	// XXX/TODO: We cannot use SS to get the latest version when lastCommitInfo
	// is nil if SS is flushed asynchronously. This is because the latest version
	// in SS might not be the latest version in the SC stores.
	//
	// Ref: https://github.com/cosmos/cosmos-sdk/issues/17314
	latestVersion, err := s.stateStore.GetLatestVersion()
	if err != nil {
		return proof.CommitID{}, err
	}

	// sanity check: ensure integrity of latest version against SC
	scVersion, err := s.stateCommitment.GetLatestVersion()
	if err != nil {
		return proof.CommitID{}, err
	}

	if scVersion != latestVersion {
		return proof.CommitID{}, fmt.Errorf("SC and SS version mismatch; got: %d, expected: %d", scVersion, latestVersion)
	}

	return proof.CommitID{Version: latestVersion}, nil
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

func (s *Store) Query(storeKey string, version uint64, key []byte, prove bool) (store.QueryResult, error) {
	if s.telemetry != nil {
		now := time.Now()
		defer s.telemetry.MeasureSince(now, "root_store", "query")
	}

	val, err := s.stateStore.Get(storeKey, version, key)
	if err != nil || val == nil {
		// fallback to querying SC backend if not found in SS backend
		//
		// Note, this should only used during migration, i.e. while SS and IAVL v2
		// are being asynchronously synced.
		if val == nil {
			bz, scErr := s.stateCommitment.Get(storeKey, version, key)
			if scErr != nil {
				return store.QueryResult{}, fmt.Errorf("failed to query SC store: %w", scErr)
			}

			val = bz
		}

		if err != nil {
			return store.QueryResult{}, fmt.Errorf("failed to query SS store: %w", err)
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

	return s.loadVersion(lv)
}

func (s *Store) LoadVersion(version uint64) error {
	if s.telemetry != nil {
		now := time.Now()
		defer s.telemetry.MeasureSince(now, "root_store", "load_version")
	}

	return s.loadVersion(version)
}

func (s *Store) loadVersion(v uint64) error {
	s.logger.Debug("loading version", "version", v)

	if err := s.stateCommitment.LoadVersion(v); err != nil {
		return fmt.Errorf("failed to load SS version %d: %w", v, err)
	}

	s.workingHash = nil
	s.commitHeader = nil

	// set lastCommitInfo explicitly s.t. Commit commits the correct version, i.e. v+1
	s.lastCommitInfo = &proof.CommitInfo{Version: v}

	return nil
}

func (s *Store) SetCommitHeader(h *coreheader.Info) {
	s.commitHeader = h
}

// WorkingHash returns the working hash of the root store. Note, WorkingHash()
// should only be called once per block once all writes are complete and prior
// to Commit() being called.
//
// If working hash is nil, then we need to compute and set it on the root store
// by constructing a CommitInfo object, which in turn creates and writes a batch
// of the current changeset to the SC tree.
func (s *Store) WorkingHash(cs *store.Changeset) ([]byte, error) {
	if s.telemetry != nil {
		now := time.Now()
		defer s.telemetry.MeasureSince(now, "root_store", "working_hash")
	}

	if s.workingHash == nil {
		if err := s.writeSC(cs); err != nil {
			return nil, err
		}

		s.workingHash = s.lastCommitInfo.Hash()
	}

	return slices.Clone(s.workingHash), nil
}

// Commit commits all state changes to the underlying SS and SC backends. Note,
// at the time of Commit(), we expect WorkingHash() to have already been called
// with the same Changeset, which internally sets the working hash, retrieved by
// writing a batch of the changeset to the SC tree, and CommitInfo on the root
// store.
func (s *Store) Commit(cs *store.Changeset) ([]byte, error) {
	if s.telemetry != nil {
		now := time.Now()
		defer s.telemetry.MeasureSince(now, "root_store", "commit")
	}

	if s.workingHash == nil {
		return nil, fmt.Errorf("working hash is nil; must call WorkingHash() before Commit()")
	}

	version := s.lastCommitInfo.Version

	if s.commitHeader != nil && uint64(s.commitHeader.Height) != version {
		s.logger.Debug("commit header and version mismatch", "header_height", s.commitHeader.Height, "version", version)
	}

	eg := new(errgroup.Group)

	// commit SS async
	eg.Go(func() error {
		if err := s.stateStore.ApplyChangeset(version, cs); err != nil {
			return fmt.Errorf("failed to commit SS: %w", err)
		}

		return nil
	})

	// commit SC async
	eg.Go(func() error {
		if err := s.commitSC(cs); err != nil {
			return fmt.Errorf("failed to commit SC: %w", err)
		}

		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	if s.commitHeader != nil {
		s.lastCommitInfo.Timestamp = s.commitHeader.Time
	}

	s.workingHash = nil

	return s.lastCommitInfo.Hash(), nil
}

// Prune prunes the root store to the provided version.
func (s *Store) Prune(version uint64) error {
	if s.telemetry != nil {
		now := time.Now()
		defer s.telemetry.MeasureSince(now, "root_store", "prune")
	}

	if err := s.stateStore.Prune(version); err != nil {
		return fmt.Errorf("failed to prune SS store: %w", err)
	}

	if err := s.stateCommitment.Prune(version); err != nil {
		return fmt.Errorf("failed to prune SC store: %w", err)
	}

	return nil
}

// writeSC accepts a Changeset and writes that as a batch to the underlying SC
// tree, which allows us to retrieve the working hash of the SC tree. Finally,
// we construct a *CommitInfo and set that as lastCommitInfo. Note, this should
// only be called once per block!
func (s *Store) writeSC(cs *store.Changeset) error {
	if err := s.stateCommitment.WriteBatch(cs); err != nil {
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
// should have already been written to the SC via WorkingHash(). This method
// solely commits that batch. An error is returned if commit fails or if the
// resulting commit hash is not equivalent to the working hash.
func (s *Store) commitSC(cs *store.Changeset) error {
	cInfo, err := s.stateCommitment.Commit(s.lastCommitInfo.Version)
	if err != nil {
		return fmt.Errorf("failed to commit SC store: %w", err)
	}

	commitHash := cInfo.Hash()

	workingHash, err := s.WorkingHash(cs)
	if err != nil {
		return fmt.Errorf("failed to get working hash: %w", err)
	}

	if !bytes.Equal(commitHash, workingHash) {
		return fmt.Errorf("unexpected commit hash; got: %X, expected: %X", commitHash, workingHash)
	}

	return nil
}
