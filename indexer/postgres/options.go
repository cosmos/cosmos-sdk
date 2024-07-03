package postgres

import (
	"cosmossdk.io/schema/logutil"
)

// Options are the options for postgres indexing.
type Options struct {
	// DisableRetainDeletions disables retain deletions functionality even on object types that have it set.
	DisableRetainDeletions bool

	// Logger is the logger for the indexer to use.
	Logger logutil.Logger
}
