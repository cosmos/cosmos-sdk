package internal

// CompactionOptions configures a single compaction run.
type CompactionOptions struct {
	// RetainVersion is the oldest version that must remain queryable after compaction.
	// Nodes orphaned before this version are eligible for removal.
	RetainVersion uint32
	// CompactionRolloverSize is the maximum size in bytes of a single compacted changeset
	// directory before a new one is started.
	CompactionRolloverSize int64
}
