package multistore

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/commitment"
	"cosmossdk.io/store/v2/types"
	"github.com/cockroachdb/errors"
	ics23 "github.com/cosmos/ics23/go"
)

type (
	// MultiStore defines an abstraction layer containing a State Storage (SS) engine
	// and one or more State Commitment (SC) engines.
	//
	// TODO:
	// - Move relevant types to the 'core' package.
	MultiStore interface {
		GetSCStore(storeKey types.StoreKey) *commitment.Database
		MountSCStore(storeKey types.StoreKey, sc *commitment.Database) error
		GetProof(storeKey types.StoreKey, version uint64, key []byte) (*ics23.CommitmentProof, error)
		LoadVersion(version uint64) error
		LoadLatestVersion() error
		GetLatestVersion() (uint64, error)
		WorkingHash() []byte
		Commit() ([]byte, error)
		SetCommitHeader(h CommitHeader)

		// TODO:
		//
		// - Tracing
		// - Branching
		// - Queries
		//
		// Ref: https://github.com/cosmos/cosmos-sdk/issues/17314

		io.Closer
	}

	CommitHeader interface {
		GetTime() time.Time
		GetHeight() uint64
	}
)

var _ MultiStore = (*Store)(nil)

type Store struct {
	logger         log.Logger
	commitHeader   CommitHeader
	initialVersion uint64

	// ss reflects the state storage backend
	ss store.VersionedDatabase

	// scStores reflect a mapping of store key to state commitment backend (i.e. a backend per module)
	scStores map[types.StoreKey]*commitment.Database

	// removalMap reflects module stores marked for removal
	removalMap map[types.StoreKey]struct{}

	// lastCommitInfo reflects the last version/hash that has been committed
	lastCommitInfo *CommitInfo

	// memListeners reflect a mapping of store key to a memory listener, which is used to flush writes to SS
	memListeners map[types.StoreKey]*types.MemoryListener
}

func New(logger log.Logger, initVersion uint64, ss store.VersionedDatabase) (MultiStore, error) {
	return &Store{
		logger:         logger.With("module", "multi_store"),
		initialVersion: initVersion,
		ss:             ss,
		scStores:       make(map[types.StoreKey]*commitment.Database),
		removalMap:     make(map[types.StoreKey]struct{}),
	}, nil
}

// Close closes the store and resets all internal fields. Note, Close() is NOT
// idempotent and should only be called once.
func (s *Store) Close() (err error) {
	err = errors.Join(err, s.ss.Close())
	for _, sc := range s.scStores {
		err = errors.Join(err, sc.Close())
	}

	s.ss = nil
	s.scStores = nil
	s.lastCommitInfo = nil
	s.commitHeader = nil
	s.removalMap = nil

	return err
}

// MountSCStore mounts a state commitment (SC) store to the multi-store. It will
// also create a new MemoryListener entry for the store key if one has not
// already been created. An error is returned if the SC store is already mounted.
func (s *Store) MountSCStore(storeKey types.StoreKey, sc *commitment.Database) error {
	s.logger.Debug("mounting store", "store_key", storeKey.String())
	if _, ok := s.scStores[storeKey]; ok {
		return fmt.Errorf("SC store with key %s already mounted", storeKey)
	}

	s.scStores[storeKey] = sc

	// Mount memory listener for the store key so we can flush accumulated writes
	// to SS upon Commit.
	if _, ok := s.memListeners[storeKey]; !ok {
		s.memListeners[storeKey] = types.NewMemoryListener()
	}

	return nil
}

// LastCommitID returns a CommitID based off of the latest internal CommitInfo.
// If an internal CommitInfo is not set, a new one will be returned with only the
// latest version set, which is based off of the SS view.
func (s *Store) LastCommitID() (CommitID, error) {
	if s.lastCommitInfo != nil {
		return s.lastCommitInfo.CommitID(), nil
	}

	// XXX/TODO: We cannot use SS to get the latest version when lastCommitInfo
	// is nil if SS is flushed asynchronously. This is because the latest version
	// in SS might not be the latest version in the SC stores.
	latestVersion, err := s.ss.GetLatestVersion()
	if err != nil {
		return CommitID{}, err
	}

	// ensure integrity of latest version across all SC stores
	for sk, sc := range s.scStores {
		scVersion := sc.GetLatestVersion()
		if scVersion != latestVersion {
			return CommitID{}, fmt.Errorf("unexpected version for %s; got: %d, expected: %d", sk, scVersion, latestVersion)
		}
	}

	return CommitID{Version: latestVersion}, nil
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

func (s *Store) GetProof(storeKey types.StoreKey, version uint64, key []byte) (*ics23.CommitmentProof, error) {
	sc, ok := s.scStores[storeKey]
	if !ok {
		return nil, fmt.Errorf("SC store with key %s not mounted", storeKey)
	}

	return sc.GetProof(version, key)
}

func (s *Store) GetSCStore(storeKey types.StoreKey) *commitment.Database {
	panic("not implemented!")
}

func (s *Store) LoadLatestVersion() error {
	lv, err := s.GetLatestVersion()
	if err != nil {
		return err
	}

	return s.loadVersion(lv, nil)
}

func (s *Store) LoadVersion(v uint64) (err error) {
	return s.loadVersion(v, nil)
}

func (s *Store) loadVersion(v uint64, upgrades any) (err error) {
	s.logger.Debug("loading version", "version", v)

	for sk, sc := range s.scStores {
		if loadErr := sc.LoadVersion(v); loadErr != nil {
			err = errors.Join(err, fmt.Errorf("failed to load version %d for %s: %w", v, sk, loadErr))
		}
	}

	// TODO: Complete this method to handle upgrades. See legacy RMS loadVersion()
	// for reference.
	//
	// Ref: https://github.com/cosmos/cosmos-sdk/issues/17314

	return err
}

func (s *Store) WorkingHash() []byte {
	storeInfos := make([]StoreInfo, 0, len(s.scStores))

	for sk, sc := range s.scStores {
		if _, ok := s.removalMap[sk]; !ok {
			storeInfos = append(storeInfos, StoreInfo{
				Name: sk.Name(),
				CommitID: CommitID{
					Hash: sc.WorkingHash(),
				},
			})
		}
	}

	sort.SliceStable(storeInfos, func(i, j int) bool {
		return storeInfos[i].Name < storeInfos[j].Name
	})

	return CommitInfo{StoreInfos: storeInfos}.Hash()
}

func (s *Store) SetCommitHeader(h CommitHeader) {
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

	// remove and close all SC stores marked for removal
	if err := s.clearSCRemovalMap(); err != nil {
		return nil, fmt.Errorf("failed to clear SC removal map: %w", err)
	}

	// commit writes to all SC stores
	commitInfo, err := s.commitSCStores(version)
	if err != nil {
		return nil, fmt.Errorf("failed to commit SC stores: %w", err)
	}

	s.lastCommitInfo = commitInfo
	s.lastCommitInfo.Timestamp = s.commitHeader.GetTime()

	// TODO: Commit writes to SS backend asynchronously.
	if err := s.commitSS(version); err != nil {
		return nil, fmt.Errorf("failed to commit SS: %w", err)
	}

	return s.lastCommitInfo.Hash(), nil
}

func (s *Store) clearSCRemovalMap() (err error) {
	for sk := range s.removalMap {
		sc, ok := s.scStores[sk]
		if ok {
			if ce := sc.Close(); ce != nil {
				err = errors.Join(err, ce)
			}

			delete(s.scStores, sk)
		}
	}

	s.removalMap = make(map[types.StoreKey]struct{})
	return err
}

// PopStateCache returns all the accumulated writes from all SC stores. Note,
// calling popStateCache destroys only the currently accumulated state in each
// listener not the state in the store itself. This is a mutating and destructive
// operation.
func (rs *Store) popStateCache() []*types.StoreKVPair {
	var writes []*types.StoreKVPair
	for _, ml := range rs.memListeners {
		if ml != nil {
			writes = append(writes, ml.PopStateCache()...)
		}
	}

	sort.SliceStable(writes, func(i, j int) bool {
		return writes[i].StoreKey < writes[j].StoreKey
	})

	return writes
}

func (s *Store) commitSS(version uint64) error {
	batch, err := s.ss.NewBatch(version)
	if err != nil {
		return err
	}

	writes := s.popStateCache()
	for _, skv := range writes {
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

// commitSCStores commits each SC store individually and returns a CommitInfo
// representing commitment of all the SC stores. Note, commitment is NOT atomic.
// An error is returned if any SC store fails to commit.
func (s *Store) commitSCStores(version uint64) (*CommitInfo, error) {
	storeInfos := make([]StoreInfo, 0, len(s.scStores))

	for sk, sc := range s.scStores {
		// TODO: Handle and support SC store last CommitID to handle the case where
		// a Commit is interrupted and a SC store could have a version that is ahead:
		//
		// Ref: https://github.com/cosmos/cosmos-sdk/issues/17314
		// scLastCommitID := sc.LastCommitID()

		// var commitID CommitID
		// if scLastCommitID.Version >= version {
		// 	scLastCommitID.Version = version
		// 	commitID = scLastCommitID
		// } else {
		// 	commitID = store.Commit()
		// }

		commitBz, err := sc.Commit()
		if err != nil {
			return nil, fmt.Errorf("failed to commit SC store %s: %w", sk, err)
		}

		storeInfos = append(storeInfos, StoreInfo{
			Name: sk.Name(),
			CommitID: CommitID{
				Version: version,
				Hash:    commitBz,
			},
		})
	}

	sort.SliceStable(storeInfos, func(i, j int) bool {
		return strings.Compare(storeInfos[i].Name, storeInfos[j].Name) < 0
	})

	return &CommitInfo{
		Version:    version,
		StoreInfos: storeInfos,
	}, nil
}
