package postgres

// Options are the options for module and object indexers.
type Options struct {
	// DisableRetainDeletions disables retain deletions functionality even on object types that have it set.
	DisableRetainDeletions bool

	// Logger is the logger for the indexer to use.
	Logger SqlLogger
}
