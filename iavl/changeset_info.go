package iavlx

type ChangesetInfo struct {
	StartVersion             uint32
	EndVersion               uint32
	LeafOrphans              uint32
	BranchOrphans            uint32
	LeafOrphanVersionTotal   uint64
	BranchOrphanVersionTotal uint64
}
