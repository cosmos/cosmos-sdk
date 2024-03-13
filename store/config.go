package store

func DefaultConfig() Config {
	return Config{
		Pruning:             "default",
		PruningKeepRecent:   "100",
		PruningInterval:     "10s",
		IAVLCacheSize:       1024,
		IAVLDisableFastNode: false,
		AppDBBackend:        "",
	}
}

// default: the last 362880 states are kept, pruning at 10 block intervals
// nothing: all historic states will be saved, nothing will be deleted (i.e. archiving node)
// everything: 2 latest states will be kept; pruning at 10 block intervals.
// custom: allow pruning options to be manually specified through 'pruning-keep-recent', and 'pruning-interval'
// pruning = "{{ .BaseConfig.Pruning }}"

// These are applied if and only if the pruning strategy is custom.

type Config struct {
	// default: the last 362880 states are kept, pruning at 10 block intervals
	// nothing: all historic states will be saved, nothing will be deleted (i.e. archiving node)
	// everything: 2 latest states will be kept; pruning at 10 block intervals.
	// custom: allow pruning options to be manually specified through 'pruning-keep-recent', and 'pruning-interval'
	Pruning string `mapstructure:"pruning" toml:"pruning"`
	// This is applied if and only if the pruning strategy is custom.
	PruningKeepRecent string `mapstructure:"pruning-keep-recent"`
	// This is applied if and only if the pruning strategy is custom.
	PruningInterval string `mapstructure:"pruning-interval"`
	//  IavlCacheSize set the size of the iavl tree cache (in number of nodes).
	// NOTE: This is only used for iavl v1.
	IAVLCacheSize uint64 `mapstructure:"iavl-cache-size" toml:"iavl-cache-size" comment:"IAVLCacheSize set the size of the iavl tree cache (in number of nodes)."`

	//  IAVLDisableFastNode enables or disables the fast node feature of IAVL.
	//  Default is false.
	// Note: This is only used for iavl v1.
	IAVLDisableFastNode bool `mapstructure:"iavl-disable-fastnode" toml:"iavl-disable-fastnode" comment:"IAVLDisableFastNode enables or disables the fast node feature of IAVL.  Default is false."`

	//  AppDBBackend defines the database backend type to use for the application and snapshots DBs.
	//  An empty string indicates that a fallback will be used.
	//  The fallback is the db_backend value set in CometBFT's config.toml.
	AppDBBackend string `mapstructure:"app-db-backend"`

	//  StateStorageBackend defines the database backend type to use for the state storage DB. Options: pebbleDB, rocksdb(cgo), sqllite(cgo).
	//  An empty string indicates that a fallback will be used.
	//  The fallback is pebbleDB.
	StateStorageBackend string `mapstructure:"state-storage-backend"`
}
