package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/client"
)

// Default constants
const (
	chainID        = ""
	keyringBackend = "os"
	output         = "text"
	node           = "tcp://localhost:26657"
	broadcastMode  = "sync"
)

type ClientConfig struct {
	ChainID        string `mapstructure:"chain-id" json:"chain-id"`
	KeyringBackend string `mapstructure:"keyring-backend" json:"keyring-backend"`
	Output         string `mapstructure:"output" json:"output"`
	Node           string `mapstructure:"node" json:"node"`
	BroadcastMode  string `mapstructure:"broadcast-mode" json:"broadcast-mode"`
}

func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{chainID, keyringBackend, output, node, broadcastMode}
}

func (c *ClientConfig) SetChainID(chainID string) {
	c.ChainID = chainID
}

func (c *ClientConfig) SetKeyringBackend(keyringBackend string) {
	c.KeyringBackend = keyringBackend
}

func (c *ClientConfig) SetOutput(output string) {
	c.Output = output
}

func (c *ClientConfig) SetNode(node string) {
	c.Node = node
}

func (c *ClientConfig) SetBroadcastMode(broadcastMode string) {
	c.BroadcastMode = broadcastMode
}

// ReadFromClientConfig reads values from client.toml file and updates them in client Context
func ReadFromClientConfig(ctx client.Context) (client.Context, error) {
	configPath := filepath.Join(ctx.HomeDir, "config")
	configFilePath := filepath.Join(configPath, "client.toml")

	conf := DefaultClientConfig()

	switch _, err := os.Stat(configFilePath); {
	// config file does not exist
	case os.IsNotExist(err):
		// we create  ~/.simapp/config/client.toml with default values

		// create a directority configPath
		if err := ensureConfigPath(configPath); err != nil {
			return ctx, fmt.Errorf("couldn't make client config: %v", err)
		}

		configTemplate, err := initConfigTemplate()
		if err != nil {
			return ctx, fmt.Errorf("could not initiate config template: %v", err)
		}

		if err := writeConfigFile(configFilePath, conf, configTemplate); err != nil {
			return ctx, fmt.Errorf("could not write client config to the file: %v", err)
		}
	// config file exists and we read config values from client.toml file
	default:
		conf, err = getClientConfig(configPath, ctx.Viper)
		if err != nil {
			return ctx, fmt.Errorf("couldn't get client config: %v", err)
		}
	}

	keyring, err := client.NewKeyringFromFlags(ctx, conf.KeyringBackend)
	if err != nil {
		return ctx, fmt.Errorf("couldn't get key ring: %v", err)
	}

	ctx = ctx.WithChainID(conf.ChainID).
		WithKeyring(keyring).
		WithOutputFormat(conf.Output).
		WithNodeURI(conf.Node).
		WithBroadcastMode(conf.BroadcastMode)

	return ctx, nil
}
