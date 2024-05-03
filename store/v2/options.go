package store

// PruneOptions defines the pruning configuration.
type PruneOptions struct {
	// KeepRecent sets the number of recent versions to keep.
	KeepRecent uint64

	// Interval sets the number of how often to prune.
	// If set to 0, no pruning will be done.
	Interval uint64
}

// DefaultPruneOptions returns the default pruning options.
// Interval is set to 0, which means no pruning will be done.
func DefaultPruneOptions() *PruneOptions {
	return &PruneOptions{
		KeepRecent: 0,
		Interval:   0,
	}
}

// ShouldPrune returns true if the given version should be pruned.
// If true, it also returns the version to prune up to.
// NOTE: The current version is not pruned.
func (opts *PruneOptions) ShouldPrune(version uint64) (bool, uint64) {
	if opts.Interval == 0 {
		return false, 0
	}

	if version <= opts.KeepRecent {
		return false, 0
	}

	if version%opts.Interval == 0 {
		return true, version - opts.KeepRecent - 1
	}

	return false, 0
}

// DBOptions defines the interface of a database options.
type DBOptions interface {
	Get(string) interface{}
}
