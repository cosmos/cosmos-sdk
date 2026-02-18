package iavl

import (
	"time"

	"github.com/cosmos/cosmos-sdk/iavl/internal"
)

type Options struct {
	// ChangesetRolloverSize is the size in bytes at which a changeset is rolled over to a new changeset.
	// By default this is set to 2GB.
	ChangesetRolloverSize int64 `json:"changeset_rollover_size"`
	// CompactionRolloverSize is the size in bytes at which a compacted changeset rolls over
	// to a new changeset.
	// By default this is set to 4GB.
	CompactionRolloverSize int64 `json:"compaction_rollover_size"`
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
	LeafEvictDepth uint8 `json:"leaf_evict_depth"`
	// CheckpointInterval is the number of versions between checkpoints.
	// By default this is set to 100.
	// Checkpoints will also always be taken when changesets are rolled over due to reaching the ChangesetRolloverSize, regardless of this setting.
	CheckpointInterval int `json:"checkpoint_interval"`
	// DisableWALFsync, if true, disables fsync calls for the WAL file.
	// This is not recommended, but may speed things up if you're using very slow storage.
	// On a modern SSD, no benchmarks have indicated that there is any commit latency as a result of fsync calls,
	// so this should not be needed.
	// CPU heavy operations generally take much longer than the WAL writing phase and WAL writing is almost never the blocker.
	DisableWALFsync bool `json:"disable_wal_fsync"`
	// RootCacheSize is the number of recent roots to cache in memory.
	// Caching recent roots can speed up access to historical state but uses more memory.
	// A recent root is defined as a recently committed version of the tree
	// or a recently checked out historical version.
	// This defaults to caching 3 recent roots. Setting this to a negative value disables root caching.
	RootCacheSize int `json:"root_cache_size"`
	// RootCacheExpiry is the expiration time in milliseconds for cached roots.
	// This defaults to 1 second.
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
		branchEvictDepth = 24 // with default evict depth 2^24 = 16M branch nodes are kept in memory
	}
	var rootCacheSize uint64 = 3 // default to caching 3 roots
	if opts.RootCacheSize > 0 {
		rootCacheSize = uint64(opts.RootCacheSize)
	} else if opts.RootCacheSize < 0 {
		rootCacheSize = 0
	}
	var rootCacheExpiry = 1 * time.Second // default to 1 second
	if opts.RootCacheExpiry > 0 {
		rootCacheExpiry = time.Duration(opts.RootCacheExpiry) * time.Millisecond
	}
	checkpointInterval := opts.CheckpointInterval
	if checkpointInterval == 0 {
		checkpointInterval = 100 // default to checkpoint every 100 versions
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
