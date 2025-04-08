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
	Type string `mapstructure:"type" toml:"type" json:"type" comment:"The name of the registered indexer type."`

	// Config are the indexer specific config options specified by the user.
	Config interface{} `mapstructure:"config" toml:"config" json:"config,omitempty" comment:"Indexer specific configuration options."`
}
