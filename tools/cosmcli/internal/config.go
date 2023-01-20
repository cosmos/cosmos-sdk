package internal

import (
	"bytes"
	"fmt"
	"os"
	"path"

	"github.com/pelletier/go-toml/v2"
	"github.com/pkg/errors"
)

type Config struct {
	Chains map[string]*ChainConfig `toml:"chains"`
}

type ChainConfig struct {
	GRPCEndpoints []GRPCEndpoint `toml:"trusted-grpc-endpoints"`
}

type GRPCEndpoint struct {
	Endpoint string `toml:"endpoint"`
	Insecure bool   `toml:"insecure"`
}

func LoadConfig(configDir string) (*Config, error) {
	configPath := configFilename(configDir)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// file doesn't exist
		return &Config{Chains: map[string]*ChainConfig{}}, nil
	}

	bz, err := os.ReadFile(configPath)
	if err != nil {
		return nil, errors.Wrapf(err, "can't read config file: %s", configPath)
	}

	config := &Config{}
	err = toml.Unmarshal(bz, config)
	if err != nil {
		return nil, errors.Wrapf(err, "can't load config file: %s", configPath)
	}

	return config, err
}

func SaveConfig(configDir string, config *Config) error {
	configPath := configFilename(configDir)
	buf := &bytes.Buffer{}
	enc := toml.NewEncoder(buf)
	err := enc.Encode(config)
	if err != nil {
		return err
	}

	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		return err
	}

	err = os.WriteFile(configPath, buf.Bytes(), 0644)
	if err != nil {
		return err
	}

	fmt.Printf("Saved config in %s\n", configPath)

	return nil
}

func configFilename(configDir string) string {
	return path.Join(configDir, "config.toml")
}
