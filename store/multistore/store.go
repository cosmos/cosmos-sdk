package multistore

import (
	"fmt"

	"cosmossdk.io/log"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/commitment"
	"cosmossdk.io/store/v2/types"
	"github.com/cockroachdb/errors"
	ics23 "github.com/cosmos/ics23/go"
)

// defaultStoreKey defines the default store key used for the single SC backend.
// Note, however, this store key is essentially irrelevant as it's not exposed
// to the user and it only needed to fulfill usage of StoreInfo during Commit.
const defaultStoreKey = "default"

var _ types.MultiStore = (*Store)(nil)

// Store defines the SDK's default MultiStore implementation. It contains a single
// State Storage (SS) backend and a single State Commitment (SC) backend. Note,
// this means all store keys are ignored and commitments exist in a single commitment
// tree.
type Store struct {
	logger         log.Logger
	initialVersion uint64

	// ss reflects the state storage backend
	ss store.VersionedDatabase

	// ss reflects the state commitment (SC) backend
	sc *commitment.Database

	// commitHeader reflects the header used when committing state (note, this isn't required and only used for query purposes)
	commitHeader types.CommitHeader

	// lastCommitInfo reflects the last version/hash that has been committed
	lastCommitInfo *types.CommitInfo

	// memListener is used to track and flush writes to the SS backend
	memListener *types.MemoryListener
}

func New(
	logger log.Logger,
	initVersion uint64,
	ss store.VersionedDatabase,
	sc *commitment.Database,
) (types.MultiStore, error) {
	return &Store{
		logger:         logger.With("module", "multi_store"),
		initialVersion: initVersion,
		ss:             ss,
		sc:             sc,
	}, nil
}

// Close closes the store and resets all internal fields. Note, Close() is NOT
// idempotent and should only be called once.
func (s *Store) Close() (err error) {
	err = errors.Join(err, s.ss.Close())
	err = errors.Join(err, s.sc.Close())

	s.ss = nil
	s.sc = nil
	s.lastCommitInfo = nil
	s.commitHeader = nil
	s.memListener = nil

	return err
}

// MountSCStore performs a no-op as a SC backend must be provided at initialization.
func (s *Store) MountSCStore(_ types.StoreKey, _ *commitment.Database) error {
	return errors.New("cannot mount SC store; SC must be provided on initialization")
}

// GetSCStore returns the store's state commitment (SC) backend. Note, the store
// key is ignored as there exists only a single SC tree.
func (s *Store) GetSCStore(_ types.StoreKey) *commitment.Database {
	return s.sc
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
func (s *Store) LastCommitID() (types.CommitID, error) {
	if s.lastCommitInfo != nil {
		return s.lastCommitInfo.CommitID(), nil
	}

	// XXX/TODO: We cannot use SS to get the latest version when lastCommitInfo
	// is nil if SS is flushed asynchronously. This is because the latest version
	// in SS might not be the latest version in the SC stores.
	//
	// Ref: https://github.com/cosmos/cosmos-sdk/issues/17314
	latestVersion, err := s.ss.GetLatestVersion()
	if err != nil {
		return types.CommitID{}, err
	}

	// ensure integrity of latest version against SC
	scVersion := s.sc.GetLatestVersion()
	if scVersion != latestVersion {
		return types.CommitID{}, fmt.Errorf("SC and SS version mismatch; got: %d, expected: %d", scVersion, latestVersion)
	}

	return types.CommitID{Version: latestVersion}, nil
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
func (s *Store) GetProof(_ types.StoreKey, version uint64, key []byte) (*ics23.CommitmentProof, error) {
	return s.sc.GetProof(version, key)
}

// LoadVersion loads a specific version returning an error upon failure.
func (s *Store) LoadVersion(v uint64) (err error) {
	return s.loadVersion(v, nil)
}

func (s *Store) GetKVStore(_ types.StoreKey) types.KVStore {
	panic("not implemented!")
}

func (s *Store) Branch() types.MultiStore {
	panic("not implemented!")
}

func (s *Store) loadVersion(v uint64, upgrades any) error {
	s.logger.Debug("loading version", "version", v)

	if err := s.sc.LoadVersion(v); err != nil {
		return fmt.Errorf("failed to load SS version %d: %w", v, err)
	}

	// TODO: Complete this method to handle upgrades. See legacy RMS loadVersion()
	// for reference.
	//
	// Ref: https://github.com/cosmos/cosmos-sdk/issues/17314

	return nil
}

func (s *Store) WorkingHash() []byte {
	storeInfos := []types.StoreInfo{
		{
			Name: defaultStoreKey,
			CommitID: types.CommitID{
				Hash: s.sc.WorkingHash(),
			},
		},
	}

	return types.CommitInfo{StoreInfos: storeInfos}.Hash()
}

func (s *Store) SetCommitHeader(h types.CommitHeader) {
	s.commitHeader = h
}

// Commit commits all state changes to the underlying SS backend and all SC stores.
// Note, writes to the SS backend are retrieved from the SC memory listeners and
// are committed in a single batch synchronously. All writes to each SC are also
// committed synchronously, however, they are NOT atomic. A byte slice is returned
// reflecting the Merkle root hash of all committed SC stores.
func (s *Store) Commit() ([]byte, error) {
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

	if s.commitHeader.GetHeight() != version {
		s.logger.Debug("commit header and version mismatch", "header_height", s.commitHeader.GetHeight(), "version", version)
	}

	// commit SC
	commitInfo, err := s.commitSC(version)
	if err != nil {
		return nil, fmt.Errorf("failed to commit SC stores: %w", err)
	}

	// commit SS
	if err := s.commitSS(version); err != nil {
		return nil, fmt.Errorf("failed to commit SS: %w", err)
	}

	s.lastCommitInfo = commitInfo
	s.lastCommitInfo.Timestamp = s.commitHeader.GetTime()

	return s.lastCommitInfo.Hash(), nil
}

// commitSC commits the SC store and returns a CommitInfo representing commitment
// of the underlying SC backend tree. Since there is only a single SC backing tree,
// all SC commits are atomic. An error is returned if commitment fails.
func (s *Store) commitSC(version uint64) (*types.CommitInfo, error) {
	commitBz, err := s.sc.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit SC store: %w", err)
	}

	storeInfos := []types.StoreInfo{
		{
			Name: defaultStoreKey,
			CommitID: types.CommitID{
				Version: version,
				Hash:    commitBz,
			},
		},
	}

	return &types.CommitInfo{
		Version:    version,
		StoreInfos: storeInfos,
	}, nil
}

// commitSS flushes all accumulated writes to the SS backend via a single batch.
// Note, this is a synchronous operation. It returns an error if the batch write
// fails.
//
// TODO: Commit writes to SS backend asynchronously.
// Ref: https://github.com/cosmos/cosmos-sdk/issues/17314
func (s *Store) commitSS(version uint64) error {
	batch, err := s.ss.NewBatch(version)
	if err != nil {
		return err
	}

	for _, skv := range s.memListener.PopStateCache() {
		if skv.Delete {
			if err := batch.Delete(skv.StoreKey, skv.Key); err != nil {
				return err
			}
		} else {
			if err := batch.Set(skv.StoreKey, skv.Key, skv.Value); err != nil {
				return err
			}
		}
	}

	return batch.Write()
}
