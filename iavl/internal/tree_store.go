package internal

type TreeStore struct{}

func (ts *TreeStore) GetChangesetForVersion(version uint32) (*ChangesetReader, Pin) {
	panic("not implemented")
}
