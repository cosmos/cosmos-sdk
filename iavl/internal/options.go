package internal

import "time"

// Options is the internal (not user facing) options type once all defaults have been filled in.
type Options struct {
	ChangesetRolloverSize  int64
	CompactionRolloverSize int64
	LeafEvictDepth         uint8
	BranchEvictDepth       uint8
	CheckpointInterval     int
	RootCacheSize          uint64
	RootCacheExpiry        time.Duration
	DisableWALFsync        bool
	DisableAutoRepair      bool
}

type TreeOptions struct {
	Options
	// ExpectedVersion is the version that we are expected to load.
	// If the actual version is ExpectedVersion + 1, we will attempt to rollback the WAL to ExpectedVersion.
	// If ExpectedVersion is zero we will load the latest version available.
	ExpectedVersion uint32
	TreeName        string
}
