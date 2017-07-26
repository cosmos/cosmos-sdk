package merkle

import (
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

func (s State) Committed() Bonsai {
	return Bonsai{s.committed}
}

func (s State) Append() Bonsai {
	return Bonsai{s.deliverTx}
}

func (s State) Check() Bonsai {
	return Bonsai{s.checkTx}
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

// Bonsai is a deformed tree forced to fit in a small pot
type Bonsai struct {
	merkle.Tree
}

// Get matches the signature of KVStore
func (b Bonsai) Get(key []byte) []byte {
	_, value, _ := b.Tree.Get(key)
	return value
}

// Set matches the signature of KVStore
func (b Bonsai) Set(key, value []byte) {
	b.Tree.Set(key, value)
}
