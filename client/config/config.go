package config

import (
	//	"fmt"
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

//	trace          = false
)

type ClientConfig struct {
	ChainID        string `mapstructure:"chain-id" json:"chain-id"`
	KeyringBackend string `mapstructure:"keyring-backend" json:"keyring-backend"`
	Output         string `mapstructure:"output" json:"output"`
	Node           string `mapstructure:"node" json:"node"`
	BroadcastMode  string `mapstructure:"broadcast-mode" json:"broadcast-mode"`
	//	Trace          bool   `mapstructure:"trace" json:"trace"`
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

/*
func (c *ClientConfig) SetTrace(trace string) error {
	boolVal, err := strconv.ParseBool(trace)
	if err != nil {
		return err
	}
	c.Trace = boolVal
	return nil
}
*/

func UpdateClientContextFromClientConfig(ctx client.Context) client.Context {
	configPath := filepath.Join(ctx.HomeDir, "config")

	/*
		if err := ensureConfigPath(configPath); err != nil {
			return ctx, fmt.Errorf("couldn't make client config: %v", err)
		}

		cliConfig, err := getClientConfig(configPath, ctx.Viper)
		if err != nil {
			return ctx, fmt.Errorf("couldn't get client config: %v", err)
		}


		keyRing, err := client.NewKeyringFromFlags(ctx, cliConfig.KeyringBackend)
		if err != nil {
			return ctx, fmt.Errorf("couldn't get key ring: %v", err)
		}
	*/

	cliConfig, _ := getClientConfig(configPath, ctx.Viper)

	keyRing, _ := client.NewKeyringFromFlags(ctx, cliConfig.KeyringBackend)

	ctx = ctx.WithChainID(cliConfig.ChainID).
		WithKeyring(keyRing).
		WithOutputFormat(cliConfig.Output).
		WithNodeURI(cliConfig.Node).
		WithBroadcastMode(cliConfig.BroadcastMode)

	return ctx
}
