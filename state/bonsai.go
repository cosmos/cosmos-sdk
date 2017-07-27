package state

import (
	"math/rand"

	"github.com/tendermint/tmlibs/merkle"
)

// store nonce as it's own type so no one can even try to fake it
type nonce int64

// Bonsai is a deformed tree forced to fit in a small pot
type Bonsai struct {
	id nonce
	merkle.Tree
}

var _ SimpleDB = &Bonsai{}

// NewBonsai wraps a merkle tree and tags it to track children
func NewBonsai(tree merkle.Tree) *Bonsai {
	return &Bonsai{
		id:   nonce(rand.Int63()),
		Tree: tree,
	}
}

// Get matches the signature of KVStore
func (b *Bonsai) Get(key []byte) []byte {
	_, value, _ := b.Tree.Get(key)
	return value
}

// Set matches the signature of KVStore
func (b *Bonsai) Set(key, value []byte) {
	b.Tree.Set(key, value)
}

func (b *Bonsai) Remove(key []byte) (value []byte) {
	value, _ = b.Tree.Remove(key)
	return
}

func (b *Bonsai) List(start, end []byte, limit int) []Model {
	res := []Model{}
	stopAtCount := func(key []byte, value []byte) (stop bool) {
		m := Model{key, value}
		res = append(res, m)
		return limit > 0 && len(res) >= limit
	}
	b.Tree.IterateRange(start, end, true, stopAtCount)
	return res
}

func (b *Bonsai) First(start, end []byte) Model {
	var m Model
	stopAtFirst := func(key []byte, value []byte) (stop bool) {
		m = Model{key, value}
		return true
	}
	b.Tree.IterateRange(start, end, true, stopAtFirst)
	return m
}

func (b *Bonsai) Last(start, end []byte) Model {
	var m Model
	stopAtFirst := func(key []byte, value []byte) (stop bool) {
		m = Model{key, value}
		return true
	}
	b.Tree.IterateRange(start, end, false, stopAtFirst)
	return m
}

func (b *Bonsai) Checkpoint() SimpleDB {
	return &Bonsai{
		id:   b.id,
		Tree: b.Tree.Copy(),
	}
}

// Commit will take all changes from the checkpoint and write
// them to the parent.
// Returns an error if this is not a child of this one
func (b *Bonsai) Commit(sub SimpleDB) error {
	bb, ok := sub.(*Bonsai)
	if !ok || (b.id != bb.id) {
		return ErrNotASubTransaction()
	}
	b.Tree = bb.Tree
	return nil
}

// Discard will remove reference to this
func (b *Bonsai) Discard() {
	b.id = 0
	b.Tree = nil
}
