package postgres

import "cosmossdk.io/schema/logutil"

// options are the options for module and object indexers.
type options struct {
	// disableRetainDeletions disables retain deletions functionality even on object types that have it set.
	disableRetainDeletions bool

	// logger is the logger for the indexer to use. It may be nil.
	logger logutil.Logger
}
