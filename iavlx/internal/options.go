package internal

import "time"

// Options is the internal (not user facing) options type once all defaults have been filled in
// by the public iavlx.Options.toInternalOpts().
type Options struct {
	// ChangesetRolloverSize is the max size in bytes before starting a new changeset directory.
	ChangesetRolloverSize int64
	// CompactionRolloverSize is the max size in bytes of a single compacted changeset directory.
	CompactionRolloverSize int64
	// LeafEvictDepth is the tree depth at which leaf nodes are evicted from the in-memory cache.
	LeafEvictDepth uint8
	// BranchEvictDepth is the tree depth at which branch nodes are evicted. Must be >= LeafEvictDepth.
	BranchEvictDepth uint8
	// CheckpointInterval is the number of versions between periodic checkpoints.
	CheckpointInterval int
	// RootCacheSize is the number of recent historical roots to cache. Zero disables caching.
	RootCacheSize uint64
	// RootCacheExpiry is how long cached historical roots are kept before eviction.
	RootCacheExpiry time.Duration
	// DisableWALFsync skips fsync on WAL writes. Not recommended for production.
	DisableWALFsync bool
	// DisableAutoRepair prevents automatic WAL truncation during startup recovery.
	DisableAutoRepair bool
}

// TreeOptions extends Options with per-tree loading configuration.
type TreeOptions struct {
	Options
	// ExpectedVersion is the version that we are expected to load.
	// If the actual version is ExpectedVersion + 1, we will attempt to rollback the WAL to ExpectedVersion.
	// If ExpectedVersion is zero we will load the latest version available.
	ExpectedVersion uint32
	TreeName        string
}
