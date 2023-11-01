package pruning

// Options defines the pruning configuration.
type Options struct {
	// KeepRecent sets the number of recent versions to keep.
	KeepRecent uint64

	// Interval sets the number of how often to prune.
	// If set to 0, no pruning will be done.
	Interval uint64

	// Sync when set to true ensure that pruning will be performed
	// synchronously, otherwise by default it will be done asynchronously.
	Sync bool
}

// DefaultOptions returns the default pruning options.
// Interval is set to 0, which means no pruning will be done.
func DefaultOptions() Options {
	return Options{
		KeepRecent: 0,
		Interval:   0,
		Sync:       false,
	}
}
