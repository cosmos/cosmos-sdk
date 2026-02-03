package iavl

type Options struct {
	ChangesetRolloverSize int   `json:"changeset_rollover_size"`
	EvictDepth            uint8 `json:"evict_depth"`
	CheckpointInterval    int   `json:"checkpoint_interval"`
	DisableWALFsync       bool  `json:"disable_wal_fsync"`
}
