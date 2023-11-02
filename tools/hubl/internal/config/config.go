package config

import (
	"bytes"
	"os"
	"path"

	"github.com/pelletier/go-toml/v2"

	"cosmossdk.io/errors"
	"cosmossdk.io/tools/hubl/internal/flags"
)

const (
	DefaultConfigDirName = ".hubl"
	GlobalKeyringDirName = "global"
)

type Config struct {
	Chains         map[string]*ChainConfig `toml:"chains"`
	KeyringBackend string                  `toml:"keyring-backend"`
}

type ChainConfig struct {
	GRPCEndpoints  []GRPCEndpoint `toml:"trusted-grpc-endpoints"`
	AddressPrefix  string         `toml:"address-prefix"`
	KeyringBackend string         `toml:"keyring-backend"`
}

type GRPCEndpoint struct {
	Endpoint string `toml:"endpoint"`
	Insecure bool   `toml:"insecure"`
}

var EmptyConfig = &Config{
	Chains:         map[string]*ChainConfig{},
	KeyringBackend: flags.DefaultKeyringBackend,
}

func (cfg *Config) GetKeyringBackend(chainName string) (string, error) {
	if chainName == GlobalKeyringDirName {
		return cfg.KeyringBackend, nil
	} else {
		chainCfg, ok := cfg.Chains[chainName]
		if ok {
			return chainCfg.KeyringBackend, nil
		}
	}

	return flags.DefaultKeyringBackend, nil
}

func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := path.Join(homeDir, DefaultConfigDirName)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return configDir, os.MkdirAll(configDir, 0o750)
	}

	return configDir, nil
}

func Load(configDir string) (*Config, error) {
	configPath := configFilename(configDir)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return EmptyConfig, nil
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

func Save(configDir string, config *Config) error {
	buf := &bytes.Buffer{}
	enc := toml.NewEncoder(buf)
	if err := enc.Encode(config); err != nil {
		return err
	}

	if err := os.MkdirAll(configDir, 0o750); err != nil {
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
