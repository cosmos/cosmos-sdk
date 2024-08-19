package postgres

import "cosmossdk.io/schema/logutil"

// options are the options for module and object indexers.
type options struct {
	// DisableRetainDeletions disables retain deletions functionality even on object types that have it set.
	DisableRetainDeletions bool

	// Logger is the logger for the indexer to use. It may be nil.
	Logger logutil.Logger
}
