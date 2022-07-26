package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Default constants
const (
	chainID        = ""
	keyringBackend = "os"
	output         = "text"
	node           = "tcp://localhost:26657"
	broadcastMode  = "sync"
	gas            = "200000"
	gasAdjustment  = "1"
	gasPrices      = ""
)

type ClientConfig struct {
	ChainID        string `mapstructure:"chain-id" json:"chain-id"`
	KeyringBackend string `mapstructure:"keyring-backend" json:"keyring-backend"`
	Output         string `mapstructure:"output" json:"output"`
	Node           string `mapstructure:"node" json:"node"`
	BroadcastMode  string `mapstructure:"broadcast-mode" json:"broadcast-mode"`
	Gas            string `mapstructure:"gas" json:"gas"`
	GasAdjustment  string `mapstructure:"gas-adjustment" json:"gas-adjustment"`
	GasPrices      string `mapstructure:"gas-prices" json:"gas-prices"`
}

// defaultClientConfig returns the reference to ClientConfig with default values.
func defaultClientConfig() *ClientConfig {
	return &ClientConfig{chainID, keyringBackend, output, node, broadcastMode, gas, gasAdjustment, gasPrices}
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

func (c *ClientConfig) SetGas(gas string) {
	c.Gas = gas
}

func (c *ClientConfig) SetGasAdjustment(gasAdj string) {
	c.GasAdjustment = gasAdj
}

func (c *ClientConfig) SetGasPrices(gasPrices string) {
	c.GasPrices = gasPrices
}

// ReadFromClientConfig reads values from client.toml file and updates them in client Context
func ReadFromClientConfig(ctx client.Context) (client.Context, error) {
	configPath := filepath.Join(ctx.HomeDir, "config")
	configFilePath := filepath.Join(configPath, "client.toml")
	conf := defaultClientConfig()

	// if config.toml file does not exist we create it and write default ClientConfig values into it.
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		if err := ensureConfigPath(configPath); err != nil {
			return ctx, fmt.Errorf("couldn't make client config: %v", err)
		}

		if ctx.ChainID != "" {
			conf.ChainID = ctx.ChainID // chain-id will be written to the client.toml while initiating the chain.
		}

		if err := writeConfigToFile(configFilePath, conf); err != nil {
			return ctx, fmt.Errorf("could not write client config to the file: %v", err)
		}
	}

	conf, err := getClientConfig(configPath, ctx.Viper)
	if err != nil {
		return ctx, fmt.Errorf("couldn't get client config: %v", err)
	}

	gasSetting, err := client.ParseGasSetting(conf.Gas)
	if err != nil {
		return ctx, fmt.Errorf("couldn't get gas setting from client config: %v", err)
	}

	ctx = ctx.WithGasSetting(gasSetting)

	gasAdj, err := strconv.ParseFloat(conf.GasAdjustment, 64)
	if err != nil {
		return ctx, fmt.Errorf("couldn't get gas adjustment from client config: %v", err)
	}

	ctx = ctx.WithGasAdjustment(gasAdj)

	gasPrices, err := sdk.ParseDecCoins(conf.GasPrices)
	if err != nil {
		return ctx, fmt.Errorf("couldn't get gas prices from client config: %v", err)
	}

	ctx = ctx.WithGasPrices(gasPrices)

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
