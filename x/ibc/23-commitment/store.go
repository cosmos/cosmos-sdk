package commitment

import (
	"bytes"
)

type Store interface {
	Prove(key, value []byte) bool
}

var _ Store = prefix{}

type prefix struct {
	store  Store
	prefix []byte
}

func NewPrefix(store Store, pref []byte) prefix {
	return prefix{
		store:  store,
		prefix: pref,
	}
}

func (prefix prefix) Prove(key, value []byte) bool {
	return prefix.store.Prove(join(prefix.prefix, key), value)
}

var _ Store = store{}

type store struct {
	root     Root
	proofs   map[string]Proof
	verified map[string][]byte
}

// Proofs must be provided
func Newstore(root Root, proofs []Proof, fullProofs []FullProof) (res store, err error) {
	res = store{
		root:     root,
		proofs:   make(map[string]Proof),
		verified: make(map[string][]byte),
	}

	for _, proof := range proofs {
		res.proofs[string(proof.Key())] = proof
	}

	for _, proof := range fullProofs {
		err = proof.Verify(root)
		if err != nil {
			return
		}
		res.verified[string(proof.Proof.Key())] = proof.Value
	}

	return
}

func (store store) Get(key []byte) ([]byte, bool) {
	res, ok := store.verified[string(key)]
	return res, ok
}

func (store store) Prove(key, value []byte) bool {
	stored, ok := store.Get(key)
	if ok && bytes.Equal(stored, value) {
		return true
	}
	proof, ok := store.proofs[string(key)]
	if !ok {
		return false
	}
	err := proof.Verify(store.root, key, value)
	if err != nil {
		return false
	}
	store.verified[string(key)] = value

	return true
}

func (store store) Proven(key []byte) bool {
	_, ok := store.Get(key)
	return ok
}
