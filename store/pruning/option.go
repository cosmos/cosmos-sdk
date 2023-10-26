package pruning

type Option struct {
	// KeepRecent sets the number of recent versions to keep.
	KeepRecent uint64

	// Interval sets the number of how often to prune.
	// If set to 0, no pruning will be done.
	Interval uint64
}

// DefaultOptions returns the default pruning options.
// Interval is set to 0, which means no pruning will be done.
func DefaultOptions() Option {
	return Option{
		KeepRecent: 0,
		Interval:   0,
	}
}
