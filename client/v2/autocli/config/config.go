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
	ChainID               string     `mapstructure:"chain-id" json:"chain-id"`
	KeyringBackend        string     `mapstructure:"keyring-backend" json:"keyring-backend"`
	KeyringDefaultKeyName string     `mapstructure:"keyring-default-keyname" json:"keyring-default-keyname"`
	Output                string     `mapstructure:"output" json:"output"`
	Node                  string     `mapstructure:"node" json:"node"`
	BroadcastMode         string     `mapstructure:"broadcast-mode" json:"broadcast-mode"`
	GRPC                  GRPCConfig `mapstructure:",squash"`
}

// GRPCConfig holds the gRPC client configuration.
type GRPCConfig struct {
	Address  string `mapstructure:"grpc-address"  json:"grpc-address"`
	Insecure bool   `mapstructure:"grpc-insecure"  json:"grpc-insecure"`
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

func CreateClientConfig(homeDir, chainID string, v *viper.Viper, customClientTemplate string, customConfig interface{}) (*Config, error) {
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

		if (customClientTemplate != "" && customConfig == nil) || (customClientTemplate == "" && customConfig != nil) {
			return nil, errors.New("customClientTemplate and customConfig should be both nil or not nil")
		}

		if customClientTemplate != "" {
			if err := setConfigTemplate(customClientTemplate); err != nil {
				return nil, fmt.Errorf("couldn't set client config template: %w", err)
			}

			if chainID != "" {
				// chain-id will be written to the client.toml while initiating the chain.
				v.Set("chain-id", chainID) // TODO: use FlagChainId
			}

			if err = v.Unmarshal(&customConfig); err != nil {
				return nil, fmt.Errorf("failed to parse custom client config: %w", err)
			}

			if err := writeConfigFile(configFilePath, customConfig); err != nil {
				return nil, fmt.Errorf("could not write client config to the file: %w", err)
			}

		} else {
			conf := DefaultConfig()
			if chainID != "" {
				// chain-id will be written to the client.toml while initiating the chain.
				conf.ChainID = chainID
			}

			if err := writeConfigFile(configFilePath, conf); err != nil {
				return nil, fmt.Errorf("could not write client config to the file: %w", err)
			}
		}
	}

	conf, err := getClientConfig(configPath, v)
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

	return CreateClientConfig(homeDir, chainID, v, "", nil)
}
