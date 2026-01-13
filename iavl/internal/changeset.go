package internal

// NOTE: This is a placeholder implementation. We will add the implementation in a future PR.

type Changeset struct {
	treeStore *TreeStore
}

func NewChangeset(treeStore *TreeStore) *Changeset {
	return &Changeset{
		treeStore: treeStore,
	}
}

//func (cs *Changeset) InitShared(files *ChangesetFiles) error {
//
//}
//
//func (cs *Changeset) Resolve(id NodeID, fileIdx uint32) (Node, Pin, error) {
//
//}

//type ChangesetReader struct {
//}
//
//func (cs *ChangesetReader) Resolve(id NodeID, fileIdx uint32) (Node, error) {
//	return nil, fmt.Errorf("not implemented")
//}
