package state

import "github.com/tendermint/iavl"

// State represents the app states, separating the commited state (for queries)
// from the working state (for CheckTx and AppendTx)
type State struct {
	committed  *Bonsai
	deliverTx  SimpleDB
	checkTx    SimpleDB
	persistent bool
}

func NewState(tree *iavl.VersionedTree) State {
	base := NewBonsai(tree)
	return State{
		committed:  base,
		deliverTx:  base.Checkpoint(),
		checkTx:    base.Checkpoint(),
		persistent: true,
	}
}

func (s State) Size() int {
	return s.committed.Tree.Size()
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

func (s State) LatestHeight() uint64 {
	return s.committed.Tree.LatestVersion()
}

func (s State) LatestHash() []byte {
	return s.committed.Tree.Hash()
}

// BatchSet is used for some weird magic in storing the new height
func (s *State) BatchSet(key, value []byte) {
	if s.persistent {
		// This is in the batch with the Save, but not in the tree
		s.committed.Tree.BatchSet(key, value)
	}
}

// Commit save persistent nodes to the database and re-copies the trees
func (s *State) Commit(version uint64) ([]byte, error) {
	// commit (if we didn't do hash earlier)
	err := s.committed.Commit(s.deliverTx)
	if err != nil {
		return nil, err
	}

	var hash []byte
	if s.persistent {
		if s.committed.Tree.Size() > 0 || s.committed.Tree.LatestVersion() > 0 {
			hash, err = s.committed.Tree.SaveVersion(version)
			if err != nil {
				return nil, err
			}
		}
	} else {
		hash = s.committed.Tree.Hash()
	}

	s.deliverTx = s.committed.Checkpoint()
	s.checkTx = s.committed.Checkpoint()
	return hash, nil
}
