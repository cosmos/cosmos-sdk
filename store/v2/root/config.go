package root

func DefaultConfig() *Config {
	return &Config{
		AppDBBackend: "goleveldb",
		Options:      DefaultStoreOptions(),
	}
}

type Config struct {
	Home         string  `toml:"-"` // this field is omitted in the TOML file
	AppDBBackend string  `mapstructure:"app-db-backend" toml:"app-db-backend" comment:"The type of database for application and snapshots databases."`
	Options      Options `mapstructure:"options" toml:"options"`
}
