package store

import (
	storev2 "cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/commitment/iavl"
	"cosmossdk.io/store/v2/root"
)

func DefaultConfig() *Config {
	return &Config{
		AppDBBackend: "",
		Options: root.Options{
			SSType: 0,
			SCType: 0,
			SCPruningOption: &storev2.PruningOption{
				KeepRecent: 2,
				Interval:   1,
			},
			SSPruningOption: &storev2.PruningOption{
				KeepRecent: 2,
				Interval:   1,
			},
			IavlConfig: &iavl.Config{
				CacheSize:              100_000,
				SkipFastStorageUpgrade: true,
			},
		},
	}
}

type Config struct {
	AppDBBackend string       `mapstructure:"app-db-backend" toml:"app-db-backend"`
	Options      root.Options `mapstructure:"options" toml:"options"`
}
