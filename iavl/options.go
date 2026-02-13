package iavl

import (
	"time"

	"github.com/cosmos/cosmos-sdk/iavl/internal"
)

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

func (opts *Options) toTreeStoreOptions() internal.TreeStoreOptions {
	changesetRolloverSize := opts.ChangesetRolloverSize
	if changesetRolloverSize == 0 {
		changesetRolloverSize = 2 * 1024 * 1024 * 1024 // 2GB default
	}
	leafEvictDepth := opts.LeafEvictDepth
	if leafEvictDepth == 0 {
		leafEvictDepth = 20 // with default evict depth 2^20 = 1M leaf nodes are kept in memory
	}
	branchEvictDepth := opts.BranchEvictDepth
	if branchEvictDepth == 0 {
		branchEvictDepth = 24
	}
	var rootCacheSize uint64 = 5 // default to caching 5 roots
	if opts.RootCacheSize > 0 {
		rootCacheSize = uint64(opts.RootCacheSize)
	} else if opts.RootCacheSize < 0 {
		rootCacheSize = 0
	}
	var rootCacheExpiry = 5 * time.Second // default to 5 seconds
	if opts.RootCacheExpiry > 0 {
		rootCacheExpiry = time.Duration(opts.RootCacheExpiry) * time.Millisecond
	}
	checkpointInterval := opts.CheckpointInterval
	if checkpointInterval == 0 {
		checkpointInterval = 1000 // default to checkpoint every 1000 versions
	}
	return internal.TreeStoreOptions{
		ChangesetRolloverSize: changesetRolloverSize,
		LeafEvictDepth:        leafEvictDepth,
		BranchEvictDepth:      branchEvictDepth,
		CheckpointInterval:    checkpointInterval,
		RootCacheSize:         rootCacheSize,
		RootCacheExpiry:       rootCacheExpiry,
	}
}
