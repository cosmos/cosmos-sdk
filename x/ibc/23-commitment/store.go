package commitment

import (
	"bytes"
)

type Store struct {
	root     Root
	proofs   map[string]Proof
	verified map[string][]byte
}

// Proofs must be provided
func NewStore(root Root, proofs []Proof, fullProofs []FullProof) (store Store, err error) {
	store = Store{
		root:     root,
		proofs:   make(map[string]Proof),
		verified: make(map[string][]byte),
	}

	for _, proof := range proofs {
		store.proofs[string(proof.Key())] = proof
	}

	for _, proof := range fullProofs {
		err = proof.Verify(root)
		if err != nil {
			return
		}
		store.verified[string(proof.Proof.Key())] = proof.Value
	}

	return
}

func (store Store) Get(key []byte) ([]byte, bool) {
	res, ok := store.verified[string(key)]
	return res, ok
}

func (store Store) Prove(key, value []byte) bool {
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

func (store Store) Proven(key []byte) bool {
	_, ok := store.Get(key)
	return ok
}
