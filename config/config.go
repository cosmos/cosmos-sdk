package config

type Config struct {
	Name		string			`mapstructure:"name"`
	CliRoot	string			`mapstructure:"homeclient"`
	Init		*InitConfig		`mapstructure:"init"`
	GenTx		*GenTxConfig	`mapstructure:"gentx"`
}

func DefaultConfig() *Config {
	return &Config{
		Name: "",
		CliRoot:	"",
		Init: DefaultInitConfig(),
		GenTx: DefaultGenTxConfig(),
	}
}

type InitConfig struct {
	ChainID		string			`mapstructure:"chainid"`
	GenTxs		bool			`mapstructure:"gentxs"`
	GenTxsDir	string			`mapstructure:"gentxsdir"`
	Overwrite	bool			`mapstructure:"overwrite"`
}

// DefaultInitConfig returns a default configuration for the `init` command
func DefaultInitConfig() *InitConfig {
	return &InitConfig{
		ChainID:	"",
		GenTxs:		false,
		GenTxsDir:	"",
		Overwrite:	false,
	}
}

type GenTxConfig struct {
	Overwrite	bool			`mapstructure:"overwrite"`
}

// DefaultGenTxConfig returns a default configuration for the `init gentx` command
func DefaultGenTxConfig() *GenTxConfig {
	return &GenTxConfig{
		Overwrite:	false,
	}
}
