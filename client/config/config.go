package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
)

func DefaultConfig() *Config {
	return &Config{
		ChainID:        "",
		KeyringBackend: "os",
		Output:         "text",
		Node:           "tcp://localhost:26657",
		BroadcastMode:  "sync",
	}
}

type Config struct {
	ChainID        string `mapstructure:"chain-id" json:"chain-id"`
	KeyringBackend string `mapstructure:"keyring-backend" json:"keyring-backend"`
	Output         string `mapstructure:"output" json:"output"`
	Node           string `mapstructure:"node" json:"node"`
	BroadcastMode  string `mapstructure:"broadcast-mode" json:"broadcast-mode"`
}

func (c *Config) SetChainID(chainID string) {
	c.ChainID = chainID
}

func (c *Config) SetKeyringBackend(keyringBackend string) {
	c.KeyringBackend = keyringBackend
}

func (c *Config) SetOutput(output string) {
	c.Output = output
}

func (c *Config) SetNode(node string) {
	c.Node = node
}

func (c *Config) SetBroadcastMode(broadcastMode string) {
	c.BroadcastMode = broadcastMode
}

// ReadFromClientConfig reads values from client.toml file and updates them in client Context
func ReadFromClientConfig(ctx client.Context, customClientTemplate string, customConfig interface{}) (client.Context, error) {
	configPath := filepath.Join(ctx.HomeDir, "config")
	configFilePath := filepath.Join(configPath, "client.toml")

	// when config.toml does not exist create and init with default values
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		if err := os.MkdirAll(configPath, os.ModePerm); err != nil {
			return ctx, fmt.Errorf("couldn't make client config: %w", err)
		}

		if customClientTemplate != "" {
			if err := setConfigTemplate(customClientTemplate); err != nil {
				return ctx, fmt.Errorf("couldn't set client config template: %w", err)
			}

			if ctx.ChainID != "" {
				// chain-id will be written to the client.toml while initiating the chain.
				ctx.Viper.Set(flags.FlagChainID, ctx.ChainID)
			}

			if err = ctx.Viper.Unmarshal(&customConfig); err != nil {
				return ctx, fmt.Errorf("failed to parse custom client config: %w", err)
			}

			if err := writeConfigFile(configFilePath, customConfig); err != nil {
				return ctx, fmt.Errorf("could not write client config to the file: %w", err)
			}

		} else {
			config, err := parseConfig(ctx.Viper)
			if err != nil {
				return ctx, fmt.Errorf("couldn't parse config: %w", err)
			}

			if ctx.ChainID != "" {
				// chain-id will be written to the client.toml while initiating the chain.
				config.ChainID = ctx.ChainID
			}

			if err := writeConfigFile(configFilePath, config); err != nil {
				return ctx, fmt.Errorf("could not write client config to the file: %w", err)
			}
		}
	}

	conf, err := getClientConfig(configPath, ctx.Viper)
	if err != nil {
		return ctx, fmt.Errorf("couldn't get client config: %w", err)
	}

	// we need to update KeyringDir field on client.Context first cause it is used in NewKeyringFromBackend
	ctx = ctx.WithOutputFormat(conf.Output).
		WithChainID(conf.ChainID).
		WithKeyringDir(ctx.HomeDir)

	keyring, err := client.NewKeyringFromBackend(ctx, conf.KeyringBackend)
	if err != nil {
		return ctx, fmt.Errorf("couldn't get keyring: %w", err)
	}

	// https://github.com/cosmos/cosmos-sdk/issues/8986
	client, err := client.NewClientFromNode(conf.Node)
	if err != nil {
		return ctx, fmt.Errorf("couldn't get client from nodeURI: %w", err)
	}

	ctx = ctx.
		WithNodeURI(conf.Node).
		WithBroadcastMode(conf.BroadcastMode).
		WithClient(client).
		WithKeyring(keyring)

	return ctx, nil
}
