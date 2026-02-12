package internal

type TreeDescription struct {
	Version                 uint32                 `json:"version"`
	RootID                  NodeID                 `json:"root_id"`
	LatestCheckpoint        uint32                 `json:"latest_checkpoint"`
	LatestSavedCheckpoint   uint32                 `json:"latest_saved_checkpoint"`
	LatestCheckpointVersion uint32                 `json:"latest_checkpoint_version"`
	TotalBytes              int                    `json:"total_bytes"`
	Changesets              []ChangesetDescription `json:"changesets"`
}

type ChangesetDescription struct {
	StartVersion  uint32           `json:"start_version"`
	EndVersion    uint32           `json:"end_version"`
	CompactedAt   uint32           `json:"compacted_at"`
	TotalLeaves   int              `json:"total_leaves"`
	TotalBranches int              `json:"total_branches"`
	TotalBytes    int              `json:"total_bytes"`
	KVLogSize     int              `json:"kv_log_size"`
	WALSize       int              `json:"wal_size"`
	Checkpoints   []CheckpointInfo `json:"checkpoints"`
	Incomplete    bool             `json:"incomplete"`
}
