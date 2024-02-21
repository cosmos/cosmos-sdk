package store

import (
	"sync/atomic"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/server/v2/core/store"
)

var _ Store = (*Storage[Database])(nil)

type Storage[SS Database] struct {
	ss SS

	// we use latest to keep track of the last height. Reading using atomics is way
	// cheaper than reading it from SS or SC everytime, it is also implementation
	// dependent.
	latest *atomic.Uint64
}

func (s Storage[SS]) StateLatest() (uint64, store.ReaderMap, error) {
	latest := s.latest.Load()
	return latest, actorsState[SS]{latest, s.ss}, nil
}

func (s Storage[SS]) StateAt(version uint64) (store.ReaderMap, error) {
	return actorsState[SS]{version, s.ss}, nil
}

func New[SS Database](ss SS) (Storage[SS], error) {
	// sanity checks.
	ssVersion, err := ss.GetLatestVersion()
	if err != nil {
		return Storage[SS]{}, err
	}

	s := Storage[SS]{
		ss:     ss,
		latest: new(atomic.Uint64),
	}
	s.latest.Store(ssVersion)
	return s, nil
}

type actorsState[SS Database] struct {
	version uint64
	ss      SS
}

func (a actorsState[SS]) GetReader(address []byte) (store.Reader, error) {
	return state[SS]{
		version:  a.version,
		storeKey: string(address),
		ss:       a.ss,
	}, nil
}

type state[SS Database] struct {
	version  uint64
	storeKey string
	ss       SS
}

func (s state[SS]) Has(key []byte) (bool, error) { return s.ss.Has(s.storeKey, s.version, key) }

func (s state[SS]) Get(bytes []byte) ([]byte, error) { return s.ss.Get(s.storeKey, s.version, bytes) }

func (s state[SS]) Iterator(start, end []byte) (corestore.Iterator, error) {
	return s.ss.Iterator(s.storeKey, s.version, start, end)
}

func (s state[SS]) ReverseIterator(start, end []byte) (corestore.Iterator, error) {
	return s.ss.ReverseIterator(s.storeKey, s.version, start, end)
}
