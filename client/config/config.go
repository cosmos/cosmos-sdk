package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/client"
)

func DefaultConfig() *ClientConfig {
	return &ClientConfig{
		ChainID:        "",
		KeyringBackend: "os",
		Output:         "text",
		Node:           "tcp://localhost:26657",
		BroadcastMode:  "sync",
	}
}

type ClientConfig struct {
	ChainID        string `mapstructure:"chain-id" json:"chain-id"`
	KeyringBackend string `mapstructure:"keyring-backend" json:"keyring-backend"`
	Output         string `mapstructure:"output" json:"output"`
	Node           string `mapstructure:"node" json:"node"`
	BroadcastMode  string `mapstructure:"broadcast-mode" json:"broadcast-mode"`
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
	homeFileDir := filepath.Join(ctx.HomeDir, "config")
	homeFilePath := filepath.Join(homeFileDir, "home.toml")
	ctx = ctx.WithHomeFilePath(homeFilePath)

	homeDir, err := ReadHomeDir(homeFileDir, ctx.Viper)
	if err == nil {
		ctx = ctx.WithHomeDir(homeDir)
	}

	configPath := filepath.Join(ctx.HomeDir, "config")
	configFilePath := filepath.Join(configPath, "client.toml")

	// when config.toml does not exist create and init with default values
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		if err := CreateNewConfigAtPath(configPath, ctx.ChainID); err != nil {
			return ctx, nil
		}
	}

	conf, err := getClientConfig(configPath, ctx.Viper)
	if err != nil {
		return ctx, fmt.Errorf("couldn't get client config: %v", err)
	}
	// we need to update KeyringDir field on Client Context first cause it is used in NewKeyringFromBackend
	ctx = ctx.WithOutputFormat(conf.Output).
		WithChainID(conf.ChainID).
		WithKeyringDir(ctx.HomeDir)

	keyring, err := client.NewKeyringFromBackend(ctx, conf.KeyringBackend)
	if err != nil {
		return ctx, fmt.Errorf("couldn't get key ring: %v", err)
	}

	ctx = ctx.WithKeyring(keyring)

	// https://github.com/cosmos/cosmos-sdk/issues/8986
	client, err := client.NewClientFromNode(conf.Node)
	if err != nil {
		return ctx, fmt.Errorf("couldn't get client from nodeURI: %v", err)
	}

	ctx = ctx.WithNodeURI(conf.Node).
		WithClient(client).
		WithBroadcastMode(conf.BroadcastMode)

	return ctx, nil
}

// CreateNewConfigAtPath sets up a basic configuration structure in the given
// directory.
func CreateNewConfigAtPath(configPath string, chainID string) error {
	configFilePath := filepath.Join(configPath, "client.toml")

	if err := os.MkdirAll(configPath, os.ModePerm); err != nil {
		return fmt.Errorf("couldn't make client config: %v", err)
	}

	conf := DefaultConfig()
	if chainID != "" {
		conf.ChainID = chainID // chain-id will be written to the client.toml while initiating the chain.
	}

	if err := writeConfigToFile(configFilePath, conf); err != nil {
		return fmt.Errorf("could not write client config to the file: %v", err)
	}

	return nil
}
