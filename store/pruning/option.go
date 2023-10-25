package pruning

type Option struct {
	// KeepRecent sets the number of recent versions to keep.
	KeepRecent uint64

	// Interval sets the number of how often to prune.
	Interval uint64
}
