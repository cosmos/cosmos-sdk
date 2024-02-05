package pruning

// PruningStore is an interface to handle pruning of SS and SC backends.
type PruningStore interface {
	// Prune attempts to prune all versions up to and including the provided version argument.
	Prune(version uint64) error
}
