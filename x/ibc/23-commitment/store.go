package commitment

import (
	"bytes"
	"errors"
)

// Store proves key-value pairs' inclusion or non-inclusion with
// the stored commitment proofs against the commitment root.
type Store interface {
	Prove(key, value []byte) bool
}

var _ Store = (*prefix)(nil) // TODO: pointer

type prefix struct {
	store  Store
	prefix []byte
}

func NewPrefix(store Store, pref []byte) Store {
	return &prefix{
		store:  store,
		prefix: pref,
	}
}

func (prefix prefix) Prove(key, value []byte) bool {
	return prefix.store.Prove(join(prefix.prefix, key), value)
}

var _ Store = (*store)(nil)

type store struct {
	root     Root
	prefix   Prefix
	proofs   map[string]Proof
	verified map[string][]byte
}

// NewStore constructs a new Store with the root, path, and proofs.
// The proofs are not proven immediately because proofs require value bytes to verify.
// If the kinds of the arguments don't match, returns error.
func NewStore(root Root, prefix Prefix, proofs []Proof) (Store, error) {
	if root.CommitmentKind() != prefix.CommitmentKind() {
		return nil, errors.New("prefix type not matching with root's")
	}

	res := &store{
		root:     root,
		prefix:   prefix,
		proofs:   make(map[string]Proof),
		verified: make(map[string][]byte),
	}

	for _, proof := range proofs {
		if proof.CommitmentKind() != root.CommitmentKind() {
			return nil, errors.New("proof type not matching with root's")
		}
		res.proofs[string(proof.GetKey())] = proof
	}

	return res, nil
}

// Get() returns the value only if it is already proven.
func (store *store) Get(key []byte) ([]byte, bool) {
	res, ok := store.verified[string(key)]
	return res, ok
}

// Prove() proves the key-value pair with the stored proof.
func (store *store) Prove(key, value []byte) bool {
	stored, ok := store.Get(key)
	if ok && bytes.Equal(stored, value) {
		return true
	}
	proof, ok := store.proofs[string(key)]
	if !ok {
		return false
	}
	err := proof.Verify(store.root, store.prefix, value)
	if err != nil {
		return false
	}
	store.verified[string(key)] = value

	return true
}

// Proven() returns true if the key-value pair is already proven
func (store *store) Proven(key []byte) bool {
	_, ok := store.Get(key)
	return ok
}
