package iavlx

type KVUpdate struct {
	SetNode   *MemNode
	DeleteKey []byte
}

type KVUpdateBatch struct {
	Version uint32
	Orphans [][]NodeID
	Updates []KVUpdate
}

func NewKVUpdateBatch(stagedVersion uint32) *KVUpdateBatch {
	return &KVUpdateBatch{
		Version: stagedVersion,
	}
}

type MutationContext struct {
	Version uint32
	Orphans []NodeID
}

func (ctx *MutationContext) MutateBranch(node Node) (*MemNode, error) {
	id := node.ID()
	if !id.IsEmpty() {
		ctx.Orphans = append(ctx.Orphans, id)
	}
	return node.MutateBranch(ctx.Version)
}

func (ctx *MutationContext) AddOrphan(id NodeID) {
	if !id.IsEmpty() {
		ctx.Orphans = append(ctx.Orphans, id)
	}
}
