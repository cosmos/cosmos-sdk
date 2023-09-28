package root

import (
	"bytes"
	"fmt"
	"slices"

	"github.com/cockroachdb/errors"
	ics23 "github.com/cosmos/ics23/go"

	"cosmossdk.io/log"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/branch"
	"cosmossdk.io/store/v2/commitment"
)

// defaultStoreKey defines the default store key used for the single SC backend.
// Note, however, this store key is essentially irrelevant as it's not exposed
// to the user and it only needed to fulfill usage of StoreInfo during Commit.
const defaultStoreKey = "default"

var _ store.RootStore = (*Store)(nil)

// Store defines the SDK's default RootStore implementation. It contains a single
// State Storage (SS) backend and a single State Commitment (SC) backend. Note,
// this means all store keys are ignored and commitments exist in a single commitment
// tree.
type Store struct {
	logger         log.Logger
	initialVersion uint64

	// stateStore reflects the state storage backend
	stateStore store.VersionedDatabase

	// stateCommitment reflects the state commitment (SC) backend
	stateCommitment *commitment.Database

	// rootKVStore reflects the root KVStore that is used to accumulate writes
	// and branch off of.
	rootKVStore store.KVStore

	// commitHeader reflects the header used when committing state (note, this isn't required and only used for query purposes)
	commitHeader store.CommitHeader

	// lastCommitInfo reflects the last version/hash that has been committed
	lastCommitInfo *store.CommitInfo

	// workingHash defines the current (yet to be committed) hash
	workingHash []byte
}

func New(
	logger log.Logger,
	initVersion uint64,
	ss store.VersionedDatabase,
	sc *commitment.Database,
) (store.RootStore, error) {
	return &Store{
		logger:          logger.With("module", "root_store"),
		initialVersion:  initVersion,
		stateStore:      ss,
		stateCommitment: sc,
		rootKVStore:     branch.New(defaultStoreKey, ss),
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

// MountSCStore performs a no-op as a SC backend must be provided at initialization.
func (s *Store) MountSCStore(_ string, _ store.Tree) error {
	return errors.New("cannot mount SC store; SC must be provided on initialization")
}

// GetSCStore returns the store's state commitment (SC) backend. Note, the store
// key is ignored as there exists only a single SC tree.
func (s *Store) GetSCStore(_ string) store.Tree {
	return s.stateCommitment
}

func (s *Store) LoadLatestVersion() error {
	lv, err := s.GetLatestVersion()
	if err != nil {
		return err
	}

	return s.loadVersion(lv, nil)
}

// LastCommitID returns a CommitID based off of the latest internal CommitInfo.
// If an internal CommitInfo is not set, a new one will be returned with only the
// latest version set, which is based off of the SS view.
func (s *Store) LastCommitID() (store.CommitID, error) {
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
		return store.CommitID{}, err
	}

	// sanity check: ensure integrity of latest version against SC
	scVersion := s.stateCommitment.GetLatestVersion()
	if scVersion != latestVersion {
		return store.CommitID{}, fmt.Errorf("SC and SS version mismatch; got: %d, expected: %d", scVersion, latestVersion)
	}

	return store.CommitID{Version: latestVersion}, nil
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

// GetProof delegates the GetProof to the store's underlying SC backend.
func (s *Store) GetProof(_ string, version uint64, key []byte) (*ics23.CommitmentProof, error) {
	return s.stateCommitment.GetProof(version, key)
}

// LoadVersion loads a specific version returning an error upon failure.
func (s *Store) LoadVersion(v uint64) (err error) {
	return s.loadVersion(v, nil)
}

// GetKVStore returns the store's root KVStore. Any writes to this store without
// branching will be committed to SC and SS upon Commit(). Branching will create
// a branched KVStore that allow writes to be discarded and propagated to the
// root KVStore using Write().
func (s *Store) GetKVStore(_ string) store.KVStore {
	return s.rootKVStore
}

func (s *Store) loadVersion(v uint64, upgrades any) error {
	s.logger.Debug("loading version", "version", v)

	if err := s.stateCommitment.LoadVersion(v); err != nil {
		return fmt.Errorf("failed to load SS version %d: %w", v, err)
	}

	// TODO: Complete this method to handle upgrades. See legacy RMS loadVersion()
	// for reference.
	//
	// Ref: https://github.com/cosmos/cosmos-sdk/issues/17314

	return nil
}

func (s *Store) SetCommitHeader(h store.CommitHeader) {
	s.commitHeader = h
}

// WorkingHash returns the working hash of the root store. Note, WorkingHash()
// should only be called once per block once all writes are complete and prior
// to Commit() being called.
//
// If working hash is nil, then we need to compute and set it on the root store
// by constructing a CommitInfo object, which in turn creates and writes a batch
// of the current changeset to the SC tree.
func (s *Store) WorkingHash() ([]byte, error) {
	if s.workingHash == nil {
		if err := s.writeSC(); err != nil {
			return nil, err
		}

		s.workingHash = s.lastCommitInfo.Hash()
	}

	return slices.Clone(s.workingHash), nil
}

// Commit commits all state changes to the underlying SS and SC backends. Note,
// at the time of Commit(), we expect WorkingHash() to have already been called,
// which internally sets the working hash, retrieved by writing a batch of the
// changeset to the SC tree, and CommitInfo on the root store. The changeset is
// retrieved from the rootKVStore and represents the entire set of writes to be
// committed. The same changeset is used to flush writes to the SS backend.
//
// Note, Commit() commits SC and SC synchronously.
func (s *Store) Commit() ([]byte, error) {
	if s.workingHash == nil {
		return nil, fmt.Errorf("working hash is nil; must call WorkingHash() before Commit()")
	}

	version := s.lastCommitInfo.Version

	if s.commitHeader.GetHeight() != version {
		s.logger.Debug("commit header and version mismatch", "header_height", s.commitHeader.GetHeight(), "version", version)
	}

	changeSet := s.rootKVStore.GetChangeSet()

	// commit SS
	if err := s.commitSS(version, changeSet); err != nil {
		return nil, fmt.Errorf("failed to commit SS: %w", err)
	}

	// commit SC
	if err := s.commitSC(); err != nil {
		return nil, fmt.Errorf("failed to commit SC stores: %w", err)
	}

	if s.commitHeader != nil {
		s.lastCommitInfo.Timestamp = s.commitHeader.GetTime()
	}

	s.rootKVStore.Reset()
	s.workingHash = nil

	return s.lastCommitInfo.Hash(), nil
}

// writeSC gets the current changeset from the rootKVStore and writes that as a
// batch to the underlying SC tree, which allows us to retrieve the working hash
// of the SC tree. Finally, we construct a *CommitInfo and return the hash.
// Note, this should only be called once per block!
func (s *Store) writeSC() error {
	changeSet := s.rootKVStore.GetChangeSet()

	if err := s.stateCommitment.WriteBatch(changeSet); err != nil {
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

	workingHash := s.stateCommitment.WorkingHash()

	s.lastCommitInfo = &store.CommitInfo{
		Version: version,
		StoreInfos: []store.StoreInfo{
			{
				Name: defaultStoreKey,
				CommitID: store.CommitID{
					Version: version,
					Hash:    workingHash,
				},
			},
		},
	}

	return nil
}

// commitSC commits the SC store. At this point, a batch of the current changeset
// should have already been written to the SC via WorkingHash(). This method
// solely commits that batch. An error is returned if commit fails or if the
// resulting commit hash is not equivalent to the working hash.
func (s *Store) commitSC() error {
	commitBz, err := s.stateCommitment.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit SC store: %w", err)
	}

	workingHash, err := s.WorkingHash()
	if err != nil {
		return fmt.Errorf("failed to get working hash: %w", err)
	}

	if bytes.Equal(commitBz, workingHash) {
		return fmt.Errorf("unexpected commit hash; got: %X, expected: %X", commitBz, workingHash)
	}

	return nil
}

// commitSS flushes all accumulated writes to the SS backend via a single batch.
// Note, this is a synchronous operation. It returns an error if the batch write
// fails.
//
// TODO: Commit writes to SS backend asynchronously.
// Ref: https://github.com/cosmos/cosmos-sdk/issues/17314
func (s *Store) commitSS(version uint64, cs *store.ChangeSet) error {
	batch, err := s.stateStore.NewBatch(version)
	if err != nil {
		return err
	}

	for _, pair := range cs.Pairs {
		if pair.Value == nil {
			if err := batch.Delete(pair.StoreKey, pair.Key); err != nil {
				return err
			}
		} else {
			if err := batch.Set(pair.StoreKey, pair.Key, pair.Value); err != nil {
				return err
			}
		}
	}

	return batch.Write()
}
