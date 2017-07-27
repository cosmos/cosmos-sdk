package state

import (
	"github.com/tendermint/merkleeyes/iavl"
	"github.com/tendermint/tmlibs/merkle"
)

// State represents the app states, separating the commited state (for queries)
// from the working state (for CheckTx and AppendTx)
type State struct {
	committed  *Bonsai
	deliverTx  *Bonsai
	checkTx    *Bonsai
	persistent bool
}

func NewState(tree merkle.Tree, persistent bool) State {
	base := NewBonsai(tree)
	return State{
		committed:  base,
		deliverTx:  base.Checkpoint().(*Bonsai),
		checkTx:    base.Checkpoint().(*Bonsai),
		persistent: persistent,
	}
}

func (s State) Committed() *Bonsai {
	return s.committed
}

func (s State) Append() *Bonsai {
	return s.deliverTx
}

func (s State) Check() *Bonsai {
	return s.checkTx
}

// Hash updates the tree
func (s *State) Hash() []byte {
	return s.deliverTx.Hash()
}

// BatchSet is used for some weird magic in storing the new height
func (s *State) BatchSet(key, value []byte) {
	if s.persistent {
		// This is in the batch with the Save, but not in the tree
		tree, ok := s.deliverTx.Tree.(*iavl.IAVLTree)
		if ok {
			tree.BatchSet(key, value)
		}
	}
}

// Commit save persistent nodes to the database and re-copies the trees
func (s *State) Commit() []byte {
	err := s.committed.Commit(s.deliverTx)
	if err != nil {
		panic(err) // ugh, TODO?
	}

	var hash []byte
	if s.persistent {
		hash = s.committed.Tree.Save()
	} else {
		hash = s.committed.Tree.Hash()
	}

	s.deliverTx = s.committed.Checkpoint().(*Bonsai)
	s.checkTx = s.committed.Checkpoint().(*Bonsai)
	return hash
}
