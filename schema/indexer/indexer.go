package indexer

import (
	"context"

	"cosmossdk.io/schema/addressutil"
	"cosmossdk.io/schema/appdata"
	"cosmossdk.io/schema/logutil"
	"cosmossdk.io/schema/view"
)

// Config species the configuration passed to an indexer initialization function.
// It includes both common configuration options related to include or excluding
// parts of the data stream as well as indexer specific options under the config
// subsection.
//
// NOTE: it is an error for an indexer to change its common options, such as adding
// or removing indexed modules, after the indexer has been initialized because this
// could result in an inconsistent state.
type Config struct {
	// Type is the name of the indexer type as registered with Register.
	Type string `json:"type"`

	// Config are the indexer specific config options specified by the user.
	Config map[string]interface{} `json:"config"`

	// ExcludeState specifies that the indexer will not receive state updates.
	ExcludeState bool `json:"exclude_state"`

	// ExcludeEvents specifies that the indexer will not receive events.
	ExcludeEvents bool `json:"exclude_events"`

	// ExcludeTxs specifies that the indexer will not receive transaction's.
	ExcludeTxs bool `json:"exclude_txs"`

	// ExcludeBlockHeaders specifies that the indexer will not receive block headers,
	// although it will still receive StartBlock and Commit callbacks, just without
	// the header data.
	ExcludeBlockHeaders bool `json:"exclude_block_headers"`

	// IncludeModules specifies a list of modules whose state the indexer will
	// receive state updates for.
	// Only one of include or exclude modules should be specified.
	IncludeModules []string `json:"include_modules"`

	// ExcludeModules specifies a list of modules whose state the indexer will not
	// receive state updates for.
	// Only one of include or exclude modules should be specified.
	ExcludeModules []string `json:"exclude_modules"`
}

type InitFunc = func(InitParams) (InitResult, error)

// InitParams is the input to the indexer initialization function.
type InitParams struct {
	// Config is the indexer config.
	Config Config

	// Context is the context that the indexer should use to listen for a shutdown signal via Context.Done(). Other
	// parameters may also be passed through context from the app if necessary. It is expected to be non-nil.
	Context context.Context

	// Logger is a logger the indexer can use to write log messages. It may be nil if the indexer does not need
	// to write logs.
	Logger logutil.Logger

	// AddressCodec is the address codec that the indexer can use to encode and decode addresses. It is
	// expected to be non-nil.
	AddressCodec addressutil.AddressCodec
}

// InitResult is the indexer initialization result and includes the indexer's listener implementation.
type InitResult struct {
	// Listener is the indexer's app data listener.
	Listener appdata.Listener

	// View is a view of indexed data that indexers can provide. It is optional and may be nil.
	// If it is provided it can be used for automated testing and other purposes.
	// At indexer start-up, the block number returned by the view will be used to determine the
	// starting block for the indexer. If the block number is 0, the indexer manager will attempt
	// to perform a catch-up sync of state. Historical events will not be replayed, but an accurate
	// representation of the current state at the height at which indexing began can be reproduced.
	// If the block number is non-zero but does not match the current chain height, a runtime error
	// will occur because this is an unsafe condition that indicates lost data.
	View view.AppData
}
