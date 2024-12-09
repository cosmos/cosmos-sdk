package config

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"cosmossdk.io/client/v2/internal/flags"
)

type Config struct {
	ChainID               string     `mapstructure:"chain-id" toml:"chain-id" comment:"The chain ID of the blockchain network"`
	KeyringBackend        string     `mapstructure:"keyring-backend" toml:"keyring-backend" comment:"The keyring backend to use (os|file|kwallet|pass|test|memory)"`
	KeyringDefaultKeyName string     `mapstructure:"keyring-default-keyname" toml:"keyring-default-keyname" comment:"The default key name to use for signing transactions"`
	Output                string     `mapstructure:"output" toml:"output" comment:"The output format for queries (text|json)"`
	Node                  string     `mapstructure:"node" toml:"node" comment:"The RPC endpoint URL for the node to connect to"`
	BroadcastMode         string     `mapstructure:"broadcast-mode" toml:"broadcast-mode" comment:"How transactions are broadcast to the network (sync|async|block)"`
	GRPC                  GRPCConfig `mapstructure:",squash" comment:"The gRPC client configuration"`
}

// GRPCConfig holds the gRPC client configuration.
type GRPCConfig struct {
	Address  string `mapstructure:"grpc-address"  toml:"grpc-address" comment:"The gRPC server address to connect to"`
	Insecure bool   `mapstructure:"grpc-insecure"  toml:"grpc-insecure" comment:"Allow gRPC over insecure connections"`
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

// CreateClientConfig creates a new client configuration or reads an existing one.
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

// CreateClientConfigFromFlags creates a client configuration from command-line flags.
func CreateClientConfigFromFlags(set *pflag.FlagSet) (*Config, error) {
	homeDir, _ := set.GetString(flags.FlagHome)
	if homeDir == "" {
		return DefaultConfig(), nil
	}
	chainID, _ := set.GetString(flags.FlagChainID)

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

// writeConfigFile renders config using the template and writes it to
// configFilePath.
func writeConfigFile(configFilePath string, config *Config) error {
	b, err := toml.Marshal(config)
	if err != nil {
		return err
	}

	if dir := filepath.Dir(configFilePath); dir != "" {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
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

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	conf := DefaultConfig()
	if err := v.Unmarshal(conf); err != nil {
		return nil, err
	}

	return conf, nil
}
