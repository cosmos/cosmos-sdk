package config

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	ChainID               string     `mapstructure:"chain-id" toml:"chain-id"`
	KeyringBackend        string     `mapstructure:"keyring-backend" toml:"keyring-backend"`
	KeyringDefaultKeyName string     `mapstructure:"keyring-default-keyname" toml:"keyring-default-keyname"`
	Output                string     `mapstructure:"output" toml:"output"`
	Node                  string     `mapstructure:"node" toml:"node"`
	BroadcastMode         string     `mapstructure:"broadcast-mode" toml:"broadcast-mode"`
	GRPC                  GRPCConfig `mapstructure:",squash"`
}

// GRPCConfig holds the gRPC client configuration.
type GRPCConfig struct {
	Address  string `mapstructure:"grpc-address"  toml:"grpc-address"`
	Insecure bool   `mapstructure:"grpc-insecure"  toml:"grpc-insecure"`
}

func DefaultConfig() *Config {
	return &Config{
		ChainID:               "",
		KeyringBackend:        "os",
		KeyringDefaultKeyName: "",
		Output:                "text",
		Node:                  "tcp://localhost:26657",
		BroadcastMode:         "sync",
	}
}

func CreateClientConfig(homeDir, chainID string, v *viper.Viper) (*Config, error) {
	if homeDir == "" {
		return nil, errors.New("home dir can't be empty")
	}

	configPath := filepath.Join(homeDir, "config")
	configFilePath := filepath.Join(configPath, "client.toml")

	// when client.toml does not exist create and init with default values
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		if err := os.MkdirAll(configPath, os.ModePerm); err != nil {
			return nil, fmt.Errorf("couldn't make client config: %w", err)
		}

		conf := DefaultConfig()
		if chainID != "" {
			// chain-id will be written to the client.toml while initiating the chain.
			conf.ChainID = chainID
		}

		if err := writeConfigFile(configFilePath, conf); err != nil {
			return nil, fmt.Errorf("could not write client config to the file: %w", err)
		}
	}

	conf, err := readConfig(configPath, v)
	if err != nil {
		return nil, fmt.Errorf("couldn't get client config: %w", err)
	}

	return conf, nil
}

func CreateClientConfigFromFlags(set *pflag.FlagSet) (*Config, error) {
	homeDir, _ := set.GetString("home")
	if homeDir == "" {
		return DefaultConfig(), nil
	}
	chainID, _ := set.GetString("chain-id")

	v := viper.New()
	executableName, err := os.Executable()
	if err != nil {
		return nil, err
	}

	v.SetEnvPrefix(path.Base(executableName))
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	return CreateClientConfig(homeDir, chainID, v)
}
