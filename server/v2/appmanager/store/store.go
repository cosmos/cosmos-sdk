package store

import (
	"sync/atomic"

	"cosmossdk.io/core/store"
)

var _ Store = (*Storage[Database])(nil)

type Storage[DB Database] struct {
	db DB

	// we use latest to keep track of the last height. Reading using atomics is way
	// cheaper than reading it from DB or SC everytime, it is also implementation
	// dependent.
	latest *atomic.Uint64
}

func (s Storage[DB]) StateLatest() (uint64, store.ReaderMap, error) {
	latest := s.latest.Load()
	return latest, actorsState[DB]{latest, s.db}, nil
}

func (s Storage[DB]) StateAt(version uint64) (store.ReaderMap, error) {
	return actorsState[DB]{version, s.db}, nil
}

func New[DB Database](db DB) (Storage[DB], error) {
	// sanity checks.
	dbVersion, err := db.GetLatestVersion()
	if err != nil {
		return Storage[DB]{}, err
	}

	s := Storage[DB]{
		db:     db,
		latest: new(atomic.Uint64),
	}
	s.latest.Store(dbVersion)
	return s, nil
}

type actorsState[DB Database] struct {
	version uint64
	db      DB
}

func (a actorsState[DB]) GetReader(address []byte) (store.Reader, error) {
	return state[DB]{
		version:  a.version,
		storeKey: address,
		db:       a.db,
	}, nil
}

type state[DB Database] struct {
	version  uint64
	storeKey []byte
	db       DB
}

func (s state[DB]) Has(key []byte) (bool, error) { return s.db.Has(s.storeKey, s.version, key) }

func (s state[DB]) Get(bytes []byte) ([]byte, error) { return s.db.Get(s.storeKey, s.version, bytes) }

func (s state[DB]) Iterator(start, end []byte) (store.Iterator, error) {
	return s.db.Iterator(s.storeKey, s.version, start, end)
}

func (s state[DB]) ReverseIterator(start, end []byte) (store.Iterator, error) {
	return s.db.ReverseIterator(s.storeKey, s.version, start, end)
}
