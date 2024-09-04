package indexer

import (
	"context"

	"cosmossdk.io/schema/addressutil"
	"cosmossdk.io/schema/appdata"
	"cosmossdk.io/schema/decoding"
	"cosmossdk.io/schema/logutil"
)

// ManagerOptions are the options for starting the indexer manager.
type ManagerOptions struct {
	// Config is the user configuration for all indexing. It should generally be an instance of map[string]interface{}
	// and match the json structure of ManagerConfig. The manager will attempt to convert it to ManagerConfig.
	Config interface{}

	// Resolver is the decoder resolver that will be used to decode the data. It is required.
	Resolver decoding.DecoderResolver

	// SyncSource is a representation of the current state of key-value data to be used in a catch-up sync.
	// Catch-up syncs will be performed at initialization when necessary. SyncSource is optional but if
	// it is omitted, indexers will only be able to start indexing state from genesis.
	SyncSource decoding.SyncSource

	// Logger is the logger that indexers can use to write logs. It is optional.
	Logger logutil.Logger

	// Context is the context that indexers should use for shutdown signals via Context.Done(). It can also
	// be used to pass down other parameters to indexers if necessary. If it is omitted, context.Background
	// will be used.
	Context context.Context

	// AddressCodec is the address codec that indexers can use to encode and decode addresses. It should always be
	// provided, but if it is omitted, the indexer manager will use a default codec which encodes and decodes addresses
	// as hex strings.
	AddressCodec addressutil.AddressCodec
}

// ManagerConfig is the configuration of the indexer manager and contains the configuration for each indexer target.
type ManagerConfig struct {
	// Target is a map of named indexer targets to their configuration.
	Target map[string]Config
}

// StartManager starts the indexer manager with the given options. The state machine should write all relevant app data to
// the returned listener.
func StartManager(opts ManagerOptions) (appdata.Listener, error) {
	panic("TODO: this will be implemented in a follow-up PR, this function is just a stub to demonstrate the API")
}
