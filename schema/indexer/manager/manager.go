package manager

import (
	"context"

	"cosmossdk.io/schema/appdata"
	"cosmossdk.io/schema/decoding"
	"cosmossdk.io/schema/logutil"
)

// Options are the options for starting the indexer manager.
type Options struct {
	// Config is the user configuration for all indexing. It should match the IndexingConfig struct and
	// the indexer manager will attempt to convert it to that data structure.
	Config interface{}

	// Resolver is the decoder resolver that will be used to decode the data.
	Resolver decoding.DecoderResolver

	// SyncSource is a representation of the current state of key-value data to be used in a catch-up sync.
	// Catch-up syncs will be performed at initialization when necessary.
	SyncSource decoding.SyncSource

	// Logger is the logger that indexers can use to write logs.
	Logger logutil.Logger

	// Context is the context that indexers should use for shutdown signals via Context.Done(). It can also
	// be used to pass down other parameters to indexers if necessary.
	Context context.Context
}

// IndexingConfig is the configuration of all the
type IndexingConfig struct {
}

// Start starts the indexer manager with the given options. The state machine should write all relevant app data to
// the returned listener.
func Start(opts Options) (appdata.Listener, error) {
	panic("TODO: this will be implemented in a follow-up PR, this function is just a stub to demonstrate the API")
}
