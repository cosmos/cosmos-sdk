package state

import "github.com/tendermint/iavl"

// State represents the app states, separating the commited state (for queries)
// from the working state (for CheckTx and AppendTx)
type State struct {
	committed   *Bonsai
	deliverTx   SimpleDB
	checkTx     SimpleDB
	historySize int64
}

// NewState wraps a versioned tree and maintains all needed
// states for the abci app
func NewState(tree *iavl.VersionedTree, historySize int64) *State {
	base := NewBonsai(tree)
	return &State{
		committed:   base,
		deliverTx:   base.Checkpoint(),
		checkTx:     base.Checkpoint(),
		historySize: historySize,
	}
}

// Size is the number of nodes in the last commit
func (s State) Size() int {
	return s.committed.Tree.Size()
}

// IsEmpty is true is no data was ever in the tree
// (and signals it is unsafe to save)
func (s State) IsEmpty() bool {
	return s.committed.Tree.IsEmpty()
}

// Committed gives us read-only access to the committed
// state(s), including historical queries
func (s State) Committed() *Bonsai {
	return s.committed
}

// Append gives us read-write access to the current working
// state (to be committed at EndBlock)
func (s State) Append() SimpleDB {
	return s.deliverTx
}

// Append gives us read-write access to the current scratch
// state (to be reset at EndBlock)
func (s State) Check() SimpleDB {
	return s.checkTx
}

// LatestHeight is the last block height we have committed
func (s State) LatestHeight() int64 {
	return int64(s.committed.Tree.Version())
}

// LatestHash is the root hash of the last state we have
// committed
func (s State) LatestHash() []byte {
	return s.committed.Tree.Hash()
}

// Commit saves persistent nodes to the database and re-copies the trees
func (s *State) Commit(version int64) ([]byte, error) {
	// commit (if we didn't do hash earlier)
	err := s.committed.Commit(s.deliverTx)
	if err != nil {
		return nil, err
	}

	// store a new version
	var hash []byte
	if !s.IsEmpty() {
		hash, _, err = s.committed.Tree.SaveVersion()
		if err != nil {
			return nil, err
		}
	}

	// release an old version
	if version > s.historySize {
		s.committed.Tree.DeleteVersion(version - s.historySize)
	}

	s.deliverTx = s.committed.Checkpoint()
	s.checkTx = s.committed.Checkpoint()
	return hash, nil
}
