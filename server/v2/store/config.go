package store

import (
	"cosmossdk.io/store/v2/root"
)

func DefaultConfig() *Config {
	return &Config{
		AppDBBackend: "goleveldb",
		Options:      root.DefaultStoreOptions(),
	}
}

type Config struct {
	AppDBBackend string       `mapstructure:"app-db-backend" toml:"app-db-backend" comment:"The type of database for application and snapshots databases."`
	Options      root.Options `mapstructure:"options" toml:"options"`
}
