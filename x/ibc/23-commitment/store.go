package commitment

import (
	"bytes"
	"errors"
)

type Store interface {
	Prove(key, value []byte) bool
	HasProof(key []byte) bool
}

var _ Store = prefix{} // TODO: pointer

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

func (prefix prefix) HasProof(key []byte) bool {
	return prefix.store.HasProof(join(prefix.prefix, key))
}

func (prefix prefix) Prove(key, value []byte) bool {
	return prefix.store.Prove(join(prefix.prefix, key), value)
}

var _ Store = (*store)(nil)

type store struct {
	root     Root
	path     Path
	proofs   map[string]Proof
	verified map[string][]byte
}

// Proofs must be provided
func NewStore(root Root, path Path, proofs []Proof) (res *store, err error) {
	if root.CommitmentKind() != path.CommitmentKind() {
		err = errors.New("path type not matching with root's")
		return
	}

	res = &store{
		root:     root,
		path:     path,
		proofs:   make(map[string]Proof),
		verified: make(map[string][]byte),
	}

	for _, proof := range proofs {
		if proof.CommitmentKind() != root.CommitmentKind() {
			err = errors.New("proof type not matching with root's")
			return
		}
		res.proofs[string(proof.GetKey())] = proof
	}

	return
}

func (store *store) Get(key []byte) ([]byte, bool) {
	res, ok := store.verified[string(key)]
	return res, ok
}

func (store *store) Prove(key, value []byte) bool {
	pathkey := store.path.Pathify(key)
	stored, ok := store.Get(pathkey)
	if ok && bytes.Equal(stored, value) {
		return true
	}
	proof, ok := store.proofs[string(pathkey)]
	if !ok {
		return false
	}
	err := proof.Verify(store.root, store.path, value)
	if err != nil {
		return false
	}
	store.verified[string(pathkey)] = value

	return true
}

func (store *store) HasProof(key []byte) bool {
	_, ok := store.proofs[string(key)]
	return ok
}

func (store *store) Proven(key []byte) bool {
	_, ok := store.Get(key)
	return ok
}
