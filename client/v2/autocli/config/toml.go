package config

import (
	"os"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/viper"
)

// writeConfigFile renders config using the template and writes it to
// configFilePath.
func writeConfigFile(configFilePath string, config *Config) error {
	b, err := toml.Marshal(config)
	if err != nil {
		return err
	}

	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		if err := os.MkdirAll(configFilePath, os.ModePerm); err != nil {
			return err
		}
	}

	return os.WriteFile(configFilePath, b, 0o600)
}

// readConfig reads values from client.toml file and unmarshalls them into ClientConfig
func readConfig(configPath string, v *viper.Viper) (*Config, error) {
	v.AddConfigPath(configPath)
	v.SetConfigName("client")
	v.SetConfigType("toml")
	v.AddConfigPath(configPath)

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	conf := DefaultConfig()
	if err := v.Unmarshal(conf); err != nil {
		return nil, err
	}

	return conf, nil
}
