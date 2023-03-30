package internal

import (
	"bytes"
	"os"
	"path"

	"cosmossdk.io/errors"
	"github.com/pelletier/go-toml/v2"
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
		return &Config{Chains: map[string]*ChainConfig{}}, nil
	}

	bz, err := os.ReadFile(configPath)
	if err != nil {
		return nil, errors.Wrapf(err, "can't read config file: %s", configPath)
	}

	config := &Config{}
	if err = toml.Unmarshal(bz, config); err != nil {
		return nil, errors.Wrapf(err, "can't load config file: %s", configPath)
	}

	return config, err
}

func SaveConfig(configDir string, config *Config) error {
	buf := &bytes.Buffer{}
	enc := toml.NewEncoder(buf)
	if err := enc.Encode(config); err != nil {
		return err
	}

	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return err
	}

	configPath := configFilename(configDir)
	if err := os.WriteFile(configPath, buf.Bytes(), 0o600); err != nil {
		return err
	}

	return nil
}

func configFilename(configDir string) string {
	return path.Join(configDir, "config.toml")
}
