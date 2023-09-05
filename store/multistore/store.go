package multistore

import (
	"fmt"
	"io"
	"sort"

	v1types "cosmossdk.io/store/types"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/commitment"
	"github.com/cockroachdb/errors"
	ics23 "github.com/cosmos/ics23/go"
)

// MultiStore defines an abstraction layer containing a State Storage (SS) engine
// and one or more State Commitment (SC) engines.
//
// TODO: Move this type to the Core package.
type MultiStore interface {
	GetSCStore(storeKey string) *commitment.Database
	MountSCStore(storeKey string, sc *commitment.Database) error
	GetProof(storeKey string, version uint64, key []byte) (*ics23.CommitmentProof, error)
	LoadVersion(version uint64) error
	GetLatestVersion() (uint64, error)
	WorkingHash() []byte
	Commit() ([]byte, error)

	// TODO:
	// - Tracing
	// - Branching
	// - Queries

	io.Closer
}

var _ MultiStore = &Store{}

// TODO:
// - Commit
// - LoadVersion

type Store struct {
	// ss reflects the state storage backend
	ss store.VersionedDatabase

	// scStores reflect a mapping of store key to state commitment backend (i.e. a backend per module)
	scStores map[string]*commitment.Database

	// removalMap reflects module stores marked for removal
	removalMap map[string]struct{}

	// lastCommitInfo reflects the last version/hash that has been committed
	lastCommitInfo *v1types.CommitInfo
}

func New(ss store.VersionedDatabase) (MultiStore, error) {
	return &Store{
		ss:         ss,
		scStores:   make(map[string]*commitment.Database),
		removalMap: make(map[string]struct{}),
	}, nil
}

func (s *Store) Close() (err error) {
	err = errors.Join(err, s.ss.Close())
	for _, sc := range s.scStores {
		err = errors.Join(err, sc.Close())
	}

	s.ss = nil
	s.scStores = nil
	s.lastCommitInfo = nil

	return err
}

func (s *Store) MountSCStore(storeKey string, sc *commitment.Database) error {
	if _, ok := s.scStores[storeKey]; ok {
		return fmt.Errorf("SC store with key %s already mounted", storeKey)
	}

	s.scStores[storeKey] = sc
	return nil
}

// LastCommitID returns the latest internal CommitID. If the latest CommitID is
// not set, a new one will be returned with the latest version set only, which
// is based off of the SS view.
func (s *Store) LastCommitID() (v1types.CommitID, error) {
	if s.lastCommitInfo == nil {
		lv, err := s.ss.GetLatestVersion()
		if err != nil {
			return v1types.CommitID{}, err
		}

		return v1types.CommitID{
			Version: int64(lv),
		}, nil
	}

	return s.lastCommitInfo.CommitID(), nil
}

// GetLatestVersion returns the latest version based on the latest internal
// CommitInfo. An error is returned if the latest CommitInfo or version cannot
// be retrieved.
func (s *Store) GetLatestVersion() (uint64, error) {
	lastCommitID, err := s.LastCommitID()
	if err != nil {
		return 0, err
	}

	return uint64(lastCommitID.Version), nil
}

func (s *Store) GetProof(storeKey string, version uint64, key []byte) (*ics23.CommitmentProof, error) {
	sc, ok := s.scStores[storeKey]
	if !ok {
		return nil, fmt.Errorf("SC store with key %s not mounted", storeKey)
	}

	return sc.GetProof(version, key)
}

func (s *Store) GetSCStore(storeKey string) *commitment.Database {
	panic("not implemented!")
}

func (s *Store) LoadVersion(v uint64) (err error) {
	for sk, sc := range s.scStores {
		if loadErr := sc.LoadVersion(v); loadErr != nil {
			err = errors.Join(err, fmt.Errorf("failed to load version %d for %s: %w", v, sk, loadErr))
		}
	}

	return err
}

func (s *Store) WorkingHash() []byte {
	storeInfos := make([]v1types.StoreInfo, 0, len(s.scStores))

	for sk, sc := range s.scStores {
		if _, ok := s.removalMap[sk]; ok {
			storeInfos = append(storeInfos, v1types.StoreInfo{
				Name: sk,
				CommitId: v1types.CommitID{
					Hash: sc.WorkingHash(),
				},
			})
		}
	}

	sort.SliceStable(storeInfos, func(i, j int) bool {
		return storeInfos[i].Name < storeInfos[j].Name
	})

	return v1types.CommitInfo{StoreInfos: storeInfos}.Hash()
}

func (s *Store) Commit() ([]byte, error) {
	panic("not implemented!")
}
