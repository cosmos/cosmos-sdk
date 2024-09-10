package indexer

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
	Config interface{} `json:"config,omitempty"`

	Filter *FilterConfig `json:"filter,omitempty"`
}

type FilterConfig struct {
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

	Modules *ModuleFilterConfig `json:"modules"`
}

type ModuleFilterConfig struct {
	// Include specifies a list of modules whose state the indexer will
	// receive state updates for.
	// Only one of include or exclude modules should be specified.
	Include []string `json:"include"`

	// Exclude specifies a list of modules whose state the indexer will not
	// receive state updates for.
	// Only one of include or exclude modules should be specified.
	Exclude []string `json:"exclude"`
}
