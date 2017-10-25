package state

import "github.com/tendermint/iavl"

// State represents the app states, separating the commited state (for queries)
// from the working state (for CheckTx and AppendTx)
type State struct {
	committed *Bonsai
	deliverTx SimpleDB
	checkTx   SimpleDB
}

func NewState(tree *iavl.VersionedTree) *State {
	base := NewBonsai(tree)
	return &State{
		committed: base,
		deliverTx: base.Checkpoint(),
		checkTx:   base.Checkpoint(),
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

// Commit save persistent nodes to the database and re-copies the trees
func (s *State) Commit(version uint64) ([]byte, error) {
	// commit (if we didn't do hash earlier)
	err := s.committed.Commit(s.deliverTx)
	if err != nil {
		return nil, err
	}

	var hash []byte
	if s.committed.Tree.Size() > 0 || s.committed.Tree.LatestVersion() > 0 {
		hash, err = s.committed.Tree.SaveVersion(version)
		if err != nil {
			return nil, err
		}
	}

	s.deliverTx = s.committed.Checkpoint()
	s.checkTx = s.committed.Checkpoint()
	return hash, nil
}
