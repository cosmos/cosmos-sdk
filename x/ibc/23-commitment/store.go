package ics23

import (
	"bytes"
	"errors"
)

// ICS 023 Function Implementation
//
// This file includes functions defined under
// https://github.com/cosmos/ics/tree/master/spec/ics-023-vector-commitments

// Store partially implements spec:verifyMembership and spec:verifyNonMembership.
// Store holds Root, Prefix, and list of Proofs that will be verified.
// Proofs incldues their respective Paths. Values are provided at the verification time.
type Store interface {
	Prove(path, value []byte) bool
}

var _ Store = prefix{}

type prefix struct {
	store  Store
	prefix []byte
}

// NewPrefix returns a prefixed store given base store and prefix.
// Prefix store for commitment proofs is used for similar path bytestring
// prefixing UX with local KVStore.
func NewPrefix(store Store, pref []byte) Store {
	return &prefix{
		store:  store,
		prefix: pref,
	}
}

// Prove implements Store.
func (prefix prefix) Prove(path, value []byte) bool {
	return prefix.store.Prove(Join(prefix.prefix, path), value)
}

var _ Store = (*store)(nil)

type store struct {
	root     Root
	prefix   Prefix
	proofs   map[string]Proof
	verified map[string][]byte
}

// NewStore constructs a new Store with the root, path, and proofs.
// The result store will be stored in the context and used by the
// commitment.Value types.
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

// Prove implements spec:verifyMembership and spec:verifyNonMembership.
// The path should be one of the path format defined under
// https://github.com/cosmos/ics/tree/master/spec/ics-024-host-requirements
// Prove retrieves the matching proof with the provided path from the internal map
// and call Verify method on it with internal Root and Prefix.
// Prove acts as verifyMembership if value is not nil, and verifyNonMembership if nil.
func (store *store) Prove(path, value []byte) bool {
	stored, ok := store.verified[string(path)]
	if ok && bytes.Equal(stored, value) {
		return true
	}
	proof, ok := store.proofs[string(path)]
	if !ok {
		return false
	}
	err := proof.Verify(store.root, store.prefix, value)
	if err != nil {
		return false
	}
	store.verified[string(path)] = value

	return true
}

func Join(a, b []byte) (res []byte) {
	res = make([]byte, len(a)+len(b))
	copy(res, a)
	copy(res[len(a):], b)
	return
}
