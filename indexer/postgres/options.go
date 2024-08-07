package postgres

import (
	"cosmossdk.io/schema/addressutil"
)

// options are the options for module and object indexers.
type options struct {
	// DisableRetainDeletions disables retain deletions functionality even on object types that have it set.
	DisableRetainDeletions bool

	// Logger is the logger for the indexer to use. It may be nil.
	Logger SqlLogger

	// AddressCodec is the codec for encoding and decoding addresses. It is expected to be non-nil.
	AddressCodec addressutil.AddressCodec
}
