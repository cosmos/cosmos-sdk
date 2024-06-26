package store

func DefaultConfig() *Config {
	return &Config{
		Pruning:           PruningOptionDefault,
		AppDBBackend:      "",
		PruningKeepRecent: 0,
		PruningInterval:   0,
	}
}

type Config struct {
	Pruning           string `mapstructure:"pruning" toml:"pruning"`
	AppDBBackend      string `mapstructure:"app-db-backend" toml:"app-db-backend"`
	PruningKeepRecent uint64 `mapstructure:"pruning-keep-recent" toml:"pruning-keep-recent"`
	PruningInterval   uint64 `mapstructure:"pruning-interval" toml:"pruning-interval"`
}
