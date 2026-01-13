package internal

type TreeStore struct {
	version uint32
	root    *NodePointer
}

func (ts *TreeStore) StagedVersion() uint32 {
	return ts.version + 1
}

func (ts *TreeStore) SaveRoot(newRoot *NodePointer) error {
	ts.root = newRoot
	ts.version++
	return nil
}

func (ts *TreeStore) Latest() *NodePointer {
	return ts.root
}
