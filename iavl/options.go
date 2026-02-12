package iavl

type Options struct {
	ChangesetRolloverSize  int `json:"changeset_rollover_size"`
	CompactionRolloverSize int `json:"compaction_rollover_size"`
	// BranchEvictDepth is the depth at which branch nodes are evicted from the cache.
	// Branch nodes occupy less memory than leaf nodes because they don't have values,
	// so it is reasonable to keep more of them in cache.
	// Must be >= LeafEvictDepth (clamped if set lower), since evicting a branch node
	// makes its children unreachable without a disk read.
	// By default this is set to 24.
	BranchEvictDepth uint8 `json:"branch_evict_depth"`
	// LeafEvictDepth is the depth at which leaf nodes are evicted from the cache.
	// Must be <= BranchEvictDepth.
	// By default this is set to 20.
	LeafEvictDepth     uint8 `json:"leaf_evict_depth"`
	CheckpointInterval int   `json:"checkpoint_interval"`
	DisableWALFsync    bool  `json:"disable_wal_fsync"`
	RootCacheSize      int   `json:"root_cache_size"`
	// RootCacheExpiry is the expiry time in milliseconds for cached roots.
	RootCacheExpiry int64 `json:"root_cache_expiry"`
}
