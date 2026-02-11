package iavl

type Options struct {
	ChangesetRolloverSize  int   `json:"changeset_rollover_size"`
	CompactionRolloverSize int   `json:"compaction_rollover_size"`
	EvictDepth             uint8 `json:"evict_depth"`
	CheckpointInterval     int   `json:"checkpoint_interval"`
	DisableWALFsync        bool  `json:"disable_wal_fsync"`
	RootCacheSize          int   `json:"root_cache_size"`
	// RootCacheExpiry is the expiry time in milliseconds for cached roots.
	RootCacheExpiry int64 `json:"root_cache_expiry"`
}
