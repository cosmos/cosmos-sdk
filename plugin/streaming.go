package plugin

import (
	"sync"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/types"
)

// STREAMING_TOML_KEY is the top-level TOML key for configuring streaming service plugins
const STREAMING_TOML_KEY = "streaming"

// GLOBAL_ACK_WAIT_LIMIT_TOML_KEY is the TOML key for configuring the global ack wait limit
const GLOBAL_ACK_WAIT_LIMIT_TOML_KEY = "global_ack_wait_limit"

// StateStreamingPlugin interface for plugins that load a baseapp.StreamingService implementation from a plugin onto a baseapp.BaseApp
type StateStreamingPlugin interface {
	// Register configures and registers the plugin streaming service with the BaseApp
	Register(bApp *baseapp.BaseApp, marshaller codec.BinaryCodec, keys map[string]*types.KVStoreKey) error

	// Start starts the background streaming process of the plugin streaming service
	Start(wg *sync.WaitGroup) error

	// Plugin is the base Plugin interface
	Plugin
}
