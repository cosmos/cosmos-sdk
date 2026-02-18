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
}

type TreeOptions struct {
	Options
	ExpectedVersion uint32
	TreeName        string
}
