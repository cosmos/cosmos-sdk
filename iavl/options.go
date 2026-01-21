package iavl

type Options struct {
	FsyncWAL              bool  `json:"fsync_wal"`
	ChangesetRolloverSize int   `json:"changeset_rollover_size"`
	EvictDepth            uint8 `json:"evict_depth"`
	CheckpointInterval    int   `json:"checkpoint_interval"`
}
