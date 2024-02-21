package store

import (
	"sync/atomic"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/server/v2/core/store"
)

var _ Store = (*Storage[Database])(nil)

type Storage[DB Database] struct {
	ss DB

	// we use latest to keep track of the last height. Reading using atomics is way
	// cheaper than reading it from DB or SC everytime, it is also implementation
	// dependent.
	latest *atomic.Uint64
}

func (s Storage[DB]) StateLatest() (uint64, store.ReaderMap, error) {
	latest := s.latest.Load()
	return latest, actorsState[DB]{latest, s.ss}, nil
}

func (s Storage[DB]) StateAt(version uint64) (store.ReaderMap, error) {
	return actorsState[DB]{version, s.ss}, nil
}

func New[DB Database](ss DB) (Storage[DB], error) {
	// sanity checks.
	ssVersion, err := ss.GetLatestVersion()
	if err != nil {
		return Storage[DB]{}, err
	}

	s := Storage[DB]{
		ss:     ss,
		latest: new(atomic.Uint64),
	}
	s.latest.Store(ssVersion)
	return s, nil
}

type actorsState[DB Database] struct {
	version uint64
	ss      DB
}

func (a actorsState[DB]) GetReader(address []byte) (store.Reader, error) {
	return state[DB]{
		version:  a.version,
		storeKey: string(address),
		ss:       a.ss,
	}, nil
}

type state[DB Database] struct {
	version  uint64
	storeKey string
	ss       DB
}

func (s state[DB]) Has(key []byte) (bool, error) { return s.ss.Has(s.storeKey, s.version, key) }

func (s state[DB]) Get(bytes []byte) ([]byte, error) { return s.ss.Get(s.storeKey, s.version, bytes) }

func (s state[DB]) Iterator(start, end []byte) (corestore.Iterator, error) {
	return s.ss.Iterator(s.storeKey, s.version, start, end)
}

func (s state[DB]) ReverseIterator(start, end []byte) (corestore.Iterator, error) {
	return s.ss.ReverseIterator(s.storeKey, s.version, start, end)
}
