package merkle

import (
	"errors"
	"math/rand"

	"github.com/tendermint/basecoin/state"
	"github.com/tendermint/merkleeyes/iavl"
	"github.com/tendermint/tmlibs/merkle"
)

// State represents the app states, separating the commited state (for queries)
// from the working state (for CheckTx and AppendTx)
type State struct {
	committed  merkle.Tree
	deliverTx  merkle.Tree
	checkTx    merkle.Tree
	persistent bool
}

func NewState(tree merkle.Tree, persistent bool) State {
	return State{
		committed:  tree,
		deliverTx:  tree.Copy(),
		checkTx:    tree.Copy(),
		persistent: persistent,
	}
}

func (s State) Committed() *Bonsai {
	return NewBonsai(s.committed)
}

func (s State) Append() *Bonsai {
	return NewBonsai(s.deliverTx)
}

func (s State) Check() *Bonsai {
	return NewBonsai(s.checkTx)
}

// Hash updates the tree
func (s *State) Hash() []byte {
	return s.deliverTx.Hash()
}

// BatchSet is used for some weird magic in storing the new height
func (s *State) BatchSet(key, value []byte) {
	if s.persistent {
		// This is in the batch with the Save, but not in the tree
		tree, ok := s.deliverTx.(*iavl.IAVLTree)
		if ok {
			tree.BatchSet(key, value)
		}
	}
}

// Commit save persistent nodes to the database and re-copies the trees
func (s *State) Commit() []byte {
	var hash []byte
	if s.persistent {
		hash = s.deliverTx.Save()
	} else {
		hash = s.deliverTx.Hash()
	}

	s.committed = s.deliverTx
	s.deliverTx = s.committed.Copy()
	s.checkTx = s.committed.Copy()
	return hash
}

// store nonce as it's own type so no one can even try to fake it
type nonce int64

// Bonsai is a deformed tree forced to fit in a small pot
type Bonsai struct {
	id nonce
	merkle.Tree
}

var _ state.SimpleDB = &Bonsai{}

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

func (b *Bonsai) List(start, end []byte, limit int) []state.Model {
	var res []state.Model
	stopAtCount := func(key []byte, value []byte) (stop bool) {
		m := state.Model{key, value}
		res = append(res, m)
		return len(res) >= limit
	}
	b.Tree.IterateRange(start, end, true, stopAtCount)
	return res
}

func (b *Bonsai) First(start, end []byte) state.Model {
	var m state.Model
	stopAtFirst := func(key []byte, value []byte) (stop bool) {
		m = state.Model{key, value}
		return true
	}
	b.Tree.IterateRange(start, end, true, stopAtFirst)
	return m
}

func (b *Bonsai) Last(start, end []byte) state.Model {
	var m state.Model
	stopAtFirst := func(key []byte, value []byte) (stop bool) {
		m = state.Model{key, value}
		return true
	}
	b.Tree.IterateRange(start, end, false, stopAtFirst)
	return m
}

func (b *Bonsai) Checkpoint() state.SimpleDB {
	return &Bonsai{
		id:   b.id,
		Tree: b.Tree.Copy(),
	}
}

// Commit will take all changes from the checkpoint and write
// them to the parent.
// Returns an error if this is not a child of this one
func (b *Bonsai) Commit(sub state.SimpleDB) error {
	bb, ok := sub.(*Bonsai)
	if !ok || (b.id != bb.id) {
		return errors.New("Not a sub-transaction")
	}
	b.Tree = bb.Tree
	return nil
}

// Discard will remove reference to this
func (b *Bonsai) Discard() {
	b.id = 0
	b.Tree = nil
}
