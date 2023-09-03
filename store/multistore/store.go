package multistore

import (
	"errors"
	"fmt"
	"io"

	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/commitment"
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
	WorkingHash() []byte
	Commit() ([]byte, error)

	// TODO:
	// - Tracing
	// - Branching
	// - Queries

	io.Closer
}

var _ MultiStore = &Store{}

type Store struct {
	ss      store.VersionedDatabase
	sc      map[string]*commitment.Database
	version uint64
}

func New(ss store.VersionedDatabase) (MultiStore, error) {
	latestVersion, err := ss.GetLatestVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest version: %w", err)
	}

	return &Store{
		ss:      ss,
		sc:      make(map[string]*commitment.Database),
		version: latestVersion,
	}, nil
}

func (s *Store) Close() (err error) {
	err = errors.Join(err, s.ss.Close())
	for _, sc := range s.sc {
		err = errors.Join(err, sc.Close())
	}

	s.ss = nil
	s.sc = nil
	s.version = 0

	return err
}

func (s *Store) MountSCStore(storeKey string, sc *commitment.Database) error {
	if _, ok := s.sc[storeKey]; ok {
		return fmt.Errorf("store with key %s already mounted", storeKey)
	}

	s.sc[storeKey] = sc
	return nil
}

func (s *Store) GetProof(storeKey string, version uint64, key []byte) (*ics23.CommitmentProof, error) {
	sc, ok := s.sc[storeKey]
	if !ok {
		return nil, fmt.Errorf("store with key %s not mounted", storeKey)
	}

	return sc.GetProof(version, key)
}
