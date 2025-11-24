package iavlx

import "time"

type Options struct {
	// EvictDepth defines the depth at which eviction occurs. 255 means no eviction.
	EvictDepth uint8 `json:"evict_depth"`

	// WriteWAL enables write-ahead logging for durability
	WriteWAL bool `json:"write_wal"`
	// CompactWAL determines if KV data is copied during compaction (true) or reused (false)
	CompactWAL bool `json:"compact_wal"`
	// DisableCompaction turns off background compaction entirely
	DisableCompaction bool `json:"disable_compaction"`

	// CompactionOrphanRatio is the orphan/total ratio (0-1) that triggers compaction
	CompactionOrphanRatio float64 `json:"compaction_orphan_ratio"`
	// CompactionOrphanAge is the average age of orphans (in versions) at which compaction is triggered
	CompactionOrphanAge uint32 `json:"compaction_orphan_age"`

	// RetainVersions is the number of recent versions to keep uncompacted.
	// If this is set to 0, all versions will be retained, and the compactor will only join changesets without removing any.
	RetainVersions uint32 `json:"retain_versions"`
	// MinCompactionSeconds is the minimum interval between compaction runs
	MinCompactionSeconds uint32 `json:"min_compaction_seconds"`
	// ChangesetMaxTarget is the maximum size of a changeset file when batching new versions
	ChangesetMaxTarget uint32 `json:"changeset_max_target"`
	// CompactionMaxTarget is the maximum size when joining/compacting old changesets
	CompactionMaxTarget uint32 `json:"compaction_max_target"`
	// CompactAfterVersions is the number of versions after which a full compaction is forced whenever there are orphans
	CompactAfterVersions uint32 `json:"compact_after_versions"`
	// ReaderUpdateInterval controls how often we create new mmap readers during batching (in versions)
	// Setting to 0 means create reader every version (high mmap churn)
	// Higher values reduce mmap overhead but delay when data becomes readable
	ReaderUpdateInterval uint32 `json:"reader_update_interval"`

	// FsyncInterval defines how often to fsync WAL when using async mode (in millisconds).
	FsyncInterval int `json:"fsync_interval"`

	// ZeroCopy attempts to reduce copying of buffers, but this isn't really implemented yet and may not even be safe to implement.
	ZeroCopy bool `json:"zero_copy"`
}

// GetCompactionOrphanAge returns the orphan age threshold with default
func (o Options) GetCompactionOrphanAge() uint32 {
	if o.CompactionOrphanAge == 0 {
		return 10 // Default to 10 versions
	}
	return o.CompactionOrphanAge
}

// GetCompactionOrphanRatio returns the orphan ratio threshold with default
func (o Options) GetCompactionOrphanRatio() float64 {
	if o.CompactionOrphanRatio <= 0 {
		return 0.6 // Default to 60% orphans
	}
	return o.CompactionOrphanRatio
}

// GetChangesetMaxTarget returns the max changeset size with default
func (o Options) GetChangesetMaxTarget() uint64 {
	if o.ChangesetMaxTarget == 0 {
		return 128 * 1024 * 1024 // 128MB default for changesets
	}
	return uint64(o.ChangesetMaxTarget)
}

// GetCompactionMaxTarget returns the max size for compaction with default
func (o Options) GetCompactionMaxTarget() uint64 {
	if o.CompactionMaxTarget == 0 {
		return 1024 * 1024 * 1024 // 1GB default for compaction
	}
	return uint64(o.CompactionMaxTarget)
}

func (o Options) GetCompactAfterVersions() uint32 {
	if o.CompactAfterVersions == 0 {
		return 500 // default to 500 versions
	}
	return o.CompactAfterVersions
}

// GetReaderUpdateInterval returns the interval for creating readers with default
func (o Options) GetReaderUpdateInterval() uint32 {
	if o.ReaderUpdateInterval == 0 {
		return 100 // Default to updating reader every 100 versions
	}
	return o.ReaderUpdateInterval
}

func (o Options) FsyncEnabled() bool {
	return o.FsyncInterval != 0
}

func (o Options) GetFsyncInterval() time.Duration {
	if o.FsyncInterval < 0 {
		return 0
	}
	return time.Millisecond * time.Duration(o.FsyncInterval)
}
