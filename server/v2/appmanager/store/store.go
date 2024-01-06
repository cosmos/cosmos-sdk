package store

import (
	"fmt"
	"sync/atomic"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/server/v2/core/store"
	storev2 "cosmossdk.io/store/v2"
)

var _ store.Store = (*Store[storev2.VersionedDatabase, storev2.Committer])(nil)

type Store[SS storev2.VersionedDatabase, SC storev2.Committer] struct {
	ss SS
	sc SC

	// we use latest to keep track of the last height. Reading using atomics is way
	// cheaper than reading it from SS or SC everytime, it is also implementation
	// dependent.
	latest *atomic.Uint64
}

func New[SS storev2.VersionedDatabase, SC storev2.Committer](ss SS, sc SC) (Store[SS, SC], error) {
	// sanity checks.
	ssVersion, err := ss.GetLatestVersion()
	if err != nil {
		return Store[SS, SC]{}, err
	}
	scVersion, err := sc.GetLatestVersion()
	if err != nil {
		return Store[SS, SC]{}, err
	}
	if scVersion != ssVersion {
		return Store[SS, SC]{}, fmt.Errorf("data corruption, sc version %d, ss version %d", scVersion, ssVersion)
	}

	s := Store[SS, SC]{
		ss:     ss,
		sc:     sc,
		latest: new(atomic.Uint64),
	}
	s.latest.Store(ssVersion)
	return s, nil
}

func (s Store[SS, SC]) StateLatest() (uint64, store.ReadonlyState, error) {
	latest := s.latest.Load()
	return latest, ssAt[SS]{latest, s.ss}, nil
}

func (s Store[SS, SC]) StateAt(version uint64) (store.ReadonlyState, error) {
	if latest := s.latest.Load(); version > latest {
		return nil, fmt.Errorf("can not create readonly state: latest %d, got: %d", latest, version)
	}
	return ssAt[SS]{version, s.ss}, nil
}

// StateCommit commits the provided state changes to SS and SC.
// NOTE: on error after applying changesets to SS we could consider
// attempting a rollback on SS.
func (s Store[SS, SC]) StateCommit(changes []store.ChangeSet) (store.Hash, error) {
	next := s.latest.Add(1)
	storeV2ChangeSet := intoStoreV2ChangeSet(changes)
	// commit ss
	err := s.ss.ApplyChangeset(next, intoStoreV2ChangeSet(changes))
	if err != nil {
		return nil, err
	}

	// commit sc, this should probably be one method only
	err = s.sc.WriteBatch(storeV2ChangeSet)
	if err != nil {
		return nil, err
	}
	si, err := s.sc.Commit()
	if err != nil {
		return nil, err
	}
	// no store keys, only one commit.
	return si[0].GetHash(), nil
}

func (s Store[SS, SC]) LatestVersion() (uint64, error) {
	return s.latest.Load(), nil
}

// ssAt implements a storev2.VersionedDatabase adapter which queries the database at the given version.
type ssAt[SS storev2.VersionedDatabase] struct {
	version uint64
	ss      SS
}

func (s ssAt[SS]) Has(key []byte) (bool, error) {
	return s.ss.Has("", s.version, key)
}

func (s ssAt[SS]) Get(bytes []byte) ([]byte, error) {
	return s.ss.Get("", s.version, bytes)
}

func (s ssAt[SS]) Iterator(start, end []byte) (corestore.Iterator, error) {
	iter, err := s.ss.Iterator("", s.version, start, end)
	if err != nil {
		return nil, err
	}
	return newIterAdapter(iter), nil
}

func (s ssAt[SS]) ReverseIterator(start, end []byte) (corestore.Iterator, error) {
	iter, err := s.ss.ReverseIterator("", s.version, start, end)
	if err != nil {
		return nil, err
	}
	return newIterAdapter(iter), nil
}
