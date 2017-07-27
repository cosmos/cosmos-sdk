package state

import (
	"github.com/tendermint/merkleeyes/iavl"
	"github.com/tendermint/tmlibs/merkle"
)

// State represents the app states, separating the commited state (for queries)
// from the working state (for CheckTx and AppendTx)
type State struct {
	committed  *Bonsai
	deliverTx  SimpleDB
	checkTx    SimpleDB
	persistent bool
}

func NewState(tree merkle.Tree, persistent bool) State {
	base := NewBonsai(tree)
	return State{
		committed:  base,
		deliverTx:  base.Checkpoint(),
		checkTx:    base.Checkpoint(),
		persistent: persistent,
	}
}

func (s State) Committed() *Bonsai {
	return s.committed
}

func (s State) Append() SimpleDB {
	return s.deliverTx
}

func (s State) Check() SimpleDB {
	return s.checkTx
}

// Hash applies deliverTx to committed and calculates hash
//
// Note the usage:
// Hash -> gets the calculated hash but doesn't save
// BatchSet -> adds some out-of-bounds data
// Commit -> Save everything to disk and the same hash
func (s *State) Hash() ([]byte, error) {
	err := s.committed.Commit(s.deliverTx)
	if err != nil {
		return nil, err
	}
	s.deliverTx = s.committed.Checkpoint()
	return s.committed.Tree.Hash(), nil
}

// BatchSet is used for some weird magic in storing the new height
func (s *State) BatchSet(key, value []byte) {
	if s.persistent {
		// This is in the batch with the Save, but not in the tree
		tree, ok := s.committed.Tree.(*iavl.IAVLTree)
		if ok {
			tree.BatchSet(key, value)
		}
	}
}

// Commit save persistent nodes to the database and re-copies the trees
func (s *State) Commit() ([]byte, error) {
	// commit (if we didn't do hash earlier)
	err := s.committed.Commit(s.deliverTx)
	if err != nil {
		return nil, err
	}

	var hash []byte
	if s.persistent {
		hash = s.committed.Tree.Save()
	} else {
		hash = s.committed.Tree.Hash()
	}

	s.deliverTx = s.committed.Checkpoint()
	s.checkTx = s.committed.Checkpoint()
	return hash, nil
}
