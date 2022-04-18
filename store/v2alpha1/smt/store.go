package smt

import (
	"crypto/sha256"
	"errors"

	dbm "github.com/cosmos/cosmos-sdk/db"
	"github.com/cosmos/cosmos-sdk/db/prefix"
	types "github.com/cosmos/cosmos-sdk/store/v2alpha1"

	ics23 "github.com/confio/ics23/go"
	"github.com/lazyledger/smt"
)

var (
	_ types.BasicKVStore = (*Store)(nil)
	_ smt.MapStore       = (dbMapStore{})
)

var (
	nodesPrefix     = []byte{0}
	preimagesPrefix = []byte{1}
	valuesPrefix    = []byte{2}

	errKeyEmpty = errors.New("key is empty or nil")
	errValueNil = errors.New("value is nil")
)

// StoreParams defines how the SMT structural and value data are accessed internally.
type StoreParams struct {
	TreeData  dbm.DBReadWriter
	ValueData dbm.DBReadWriter
}

// Store Implements types.BasicKVStore.
type Store struct {
	tree smt.SparseMerkleTree
	// Maps value hash back to preimage
	values dbm.DBReadWriter
	// Maps hashed key (tree path) back to preimage
	preimages dbm.DBReadWriter
}

// An smt.MapStore that wraps Get to raise smt.InvalidKeyError;
// smt.SparseMerkleTree expects this error to be returned when a key is not found
type dbMapStore struct{ dbm.DBReadWriter }

func NewStore(par StoreParams) *Store {
	nodes := prefix.NewPrefixReadWriter(par.TreeData, nodesPrefix)
	preimages := prefix.NewPrefixReadWriter(par.TreeData, preimagesPrefix)
	values := par.ValueData
	if values == nil {
		values = prefix.NewPrefixReadWriter(par.TreeData, valuesPrefix)
	}
	return &Store{
		tree:      smt.NewSMT(dbMapStore{nodes}, sha256.New()),
		values:    values,
		preimages: preimages,
	}
}

func LoadStore(par StoreParams, root []byte) *Store {
	nodes := prefix.NewPrefixReadWriter(par.TreeData, nodesPrefix)
	preimages := prefix.NewPrefixReadWriter(par.TreeData, preimagesPrefix)
	values := par.ValueData
	if values == nil {
		values = prefix.NewPrefixReadWriter(par.TreeData, valuesPrefix)
	}
	return &Store{
		tree:      smt.ImportSMT(dbMapStore{nodes}, sha256.New(), root),
		values:    values,
		preimages: preimages,
	}
}

func (s *Store) GetProofICS23(key []byte) (*ics23.CommitmentProof, error) {
	return createIcs23Proof(s, key)
}

func (s *Store) Root() []byte { return s.tree.Root() }

// BasicKVStore interface below:

// Get returns nil iff key doesn't exist. Panics on nil or empty key.
func (s *Store) Get(key []byte) []byte {
	if len(key) == 0 {
		panic(errKeyEmpty)
	}
	val, err := s.values.Get(key)
	if err != nil {
		panic(err)
	}
	return val
}

// Has checks if a key exists. Panics on nil or empty key.
func (s *Store) Has(key []byte) bool {
	if len(key) == 0 {
		panic(errKeyEmpty)
	}
	has, err := s.values.Has(key)
	if err != nil {
		panic(err)
	}
	return has
}

// Set sets the key. Panics on nil key or value.
func (s *Store) Set(key []byte, value []byte) {
	if len(key) == 0 {
		panic(errKeyEmpty)
	}
	if value == nil {
		panic(errValueNil)
	}
	if err := s.tree.Update(key, value); err != nil {
		panic(err)
	}
	if err := s.values.Set(key, value); err != nil {
		panic(err)
	}
	// TODO: plug into the SMT's hashers
	path := sha256.Sum256(key)
	if err := s.preimages.Set(path[:], key); err != nil {
		panic(err)
	}
}

// Delete deletes the key. Panics on nil key.
func (s *Store) Delete(key []byte) {
	if len(key) == 0 {
		panic(errKeyEmpty)
	}
	if err := s.tree.Delete(key); err != nil && err != smt.ErrKeyNotPresent {
		panic(err)
	}
	if err := s.values.Delete(key); err != nil {
		panic(err)
	}
	path := sha256.Sum256(key)
	s.preimages.Delete(path[:])
}

func (s *Store) Commit() error {
	return s.tree.Save()
}

func (ms dbMapStore) Get(key []byte) ([]byte, error) {
	val, err := ms.DBReadWriter.Get(key)
	if err != nil {
		return nil, err
	}
	if val == nil {
		return nil, &smt.InvalidKeyError{key}
	}
	return val, nil
}

func (ms dbMapStore) Delete(key []byte) error {
	has, err := ms.DBReadWriter.Has(key)
	if err != nil {
		return err
	}
	if !has {
		return &smt.InvalidKeyError{key}
	}
	return ms.DBReadWriter.Delete(key)
}
