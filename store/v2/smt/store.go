package smt

import (
	"crypto/sha256"
	"errors"

	dbm "github.com/cosmos/cosmos-sdk/db"
	"github.com/cosmos/cosmos-sdk/store/types"

	"github.com/lazyledger/smt"
	tmcrypto "github.com/tendermint/tendermint/proto/tendermint/crypto"
)

var (
	_ types.BasicKVStore = (*Store)(nil)
	_ smt.MapStore       = (dbMapStore{})
)

var (
	errKeyEmpty = errors.New("key is empty or nil")
	errValueNil = errors.New("value is nil")
)

// Store Implements types.KVStore and CommitKVStore.
type Store struct {
	tree *smt.SparseMerkleTree
}

// An smt.MapStore that wraps Get to raise smt.InvalidKeyError;
// smt.SparseMerkleTree expects this error to be returned when a key is not found
type dbMapStore struct{ dbm.DBReadWriter }

func NewStore(nodes, values dbm.DBReadWriter) *Store {
	return &Store{
		tree: smt.NewSparseMerkleTree(dbMapStore{nodes}, dbMapStore{values}, sha256.New()),
	}
}

func LoadStore(nodes, values dbm.DBReadWriter, root []byte) *Store {
	return &Store{
		tree: smt.ImportSparseMerkleTree(dbMapStore{nodes}, dbMapStore{values}, sha256.New(), root),
	}
}

func (s *Store) GetProof(key []byte) (*tmcrypto.ProofOps, error) {
	proof, err := s.tree.Prove(key)
	if err != nil {
		return nil, err
	}
	op := NewProofOp(s.tree.Root(), key, SHA256, proof)
	return &tmcrypto.ProofOps{Ops: []tmcrypto.ProofOp{op.ProofOp()}}, nil
}

func (s *Store) Root() []byte { return s.tree.Root() }

// BasicKVStore interface below:

// Get returns nil iff key doesn't exist. Panics on nil key.
func (s *Store) Get(key []byte) []byte {
	if len(key) == 0 {
		panic(errKeyEmpty)
	}
	val, err := s.tree.Get(key)
	if err != nil {
		panic(err)
	}
	return val
}

// Has checks if a key exists. Panics on nil key.
func (s *Store) Has(key []byte) bool {
	if len(key) == 0 {
		panic(errKeyEmpty)
	}
	has, err := s.tree.Has(key)
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
	_, err := s.tree.Update(key, value)
	if err != nil {
		panic(err)
	}
}

// Delete deletes the key. Panics on nil key.
func (s *Store) Delete(key []byte) {
	if len(key) == 0 {
		panic(errKeyEmpty)
	}
	_, err := s.tree.Delete(key)
	if err != nil {
		panic(err)
	}
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
