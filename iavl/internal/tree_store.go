package internal

type TreeStore struct {
	stagedVersion uint32
	root          *NodePointer
}

func (ts *TreeStore) StagedVersion() uint32 {
	return ts.stagedVersion
}

func (ts *TreeStore) SaveRoot(newRoot *NodePointer) error {
	ts.root = newRoot
	return nil
}
