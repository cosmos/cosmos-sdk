package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"

	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
)

// Default constants
const (
	chainID        = ""
	keyringBackend = "os"
	output         = "text"
	node           = "tcp://localhost:26657"
	broadcastMode  = "sync"
	trace          = false
)

var ErrWrongNumberOfArgs = fmt.Errorf("wrong number of arguments")

type ClientConfig struct {
	ChainID        string `mapstructure:"chain-id" json:"chain-id"`
	KeyringBackend string `mapstructure:"keyring-backend" json:"keyring-backend"`
	Output         string `mapstructure:"output" json:"output"`
	Node           string `mapstructure:"node" json:"node"`
	BroadcastMode  string `mapstructure:"broadcast-mode" json:"broadcast-mode"`
	Trace          bool   `mapstructure:"trace" json:"trace"`
}

// TODO Validate values in setters
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

func (c *ClientConfig) SetTrace(trace string) error {
	boolVal, err := strconv.ParseBool(trace)
	if err != nil {
		return err
	}
	c.Trace = boolVal
	return nil
}

func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{chainID, keyringBackend, output, node, broadcastMode, trace}
}

// Cmd returns a CLI command to interactively create an application CLI
// config file.
func Cmd(defaultCLIHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config <key> [value]",
		Short: "Create or query an application CLI configuration file",
		RunE:  runConfigCmd,
		Args:  cobra.RangeArgs(0, 2),
	}

	cmd.Flags().String(flags.FlagHome, defaultCLIHome,
		"set client's home directory for configuration")
	return cmd
}

func runConfigCmd(cmd *cobra.Command, args []string) error {

	clientCtx := client.GetClientContextFromCmd(cmd)

	configPath := filepath.Join(clientCtx.HomeDir, "config")

	if err := ensureConfigPath(configPath); err != nil {
		return fmt.Errorf("couldn't make config config: %v", err)
	}

	cliConfig, err := getClientConfig(configPath, clientCtx.Viper)
	if err != nil {
		return fmt.Errorf("couldn't get client config: %v", err)
	}

	switch len(args) {
	case 0:
		// print all client config fields to sdt out
		s, _ := json.MarshalIndent(cliConfig, "", "\t")
		cmd.Println(string(s))

	case 1:
		// it's a get
		// TODO implement method for get
		// should i implement getters here?
		key := args[0]
		switch key {
		case flags.FlagChainID:
			cmd.Println(cliConfig.ChainID)
		case flags.FlagKeyringBackend:
			cmd.Println(cliConfig.KeyringBackend)
		case tmcli.OutputFlag:
			cmd.Println(cliConfig.Output)
		case flags.FlagNode:
			cmd.Println(cliConfig.Node)
		case flags.FlagBroadcastMode:
			cmd.Println(cliConfig.BroadcastMode)
		case "trace":
			cmd.Println(cliConfig.Trace)
		default:
			err := errUnknownConfigKey(key)
			return fmt.Errorf("couldn't get the value for the key: %v, error:  %v\n", key, err)
		}

	case 2:
		// it's set

		key, value := args[0], args[1]

		switch key {
		case flags.FlagChainID:
			cliConfig.SetChainID(value)
		case flags.FlagKeyringBackend:
			cliConfig.SetKeyringBackend(value)
		case tmcli.OutputFlag:
			cliConfig.SetOutput(value)
		case flags.FlagNode:
			cliConfig.SetNode(value)
		case flags.FlagBroadcastMode:
			cliConfig.SetBroadcastMode(value)
		case "trace":
			if err := cliConfig.SetTrace(value); err != nil {
				return fmt.Errorf("couldn't parse bool value: %v", err)
			}
		default:
			return errUnknownConfigKey(key)
		}

		configTemplate, err := InitConfigTemplate()
		if err != nil {
			return fmt.Errorf("could not initiate config template: %v", err)
		}

		cliConfigFile := filepath.Join(configPath, "client.toml")
		if err := WriteConfigFile(cliConfigFile, cliConfig, configTemplate); err != nil {
			return fmt.Errorf("could not write client config to the file: %v", err)
		}

	default:
		// print error
		return errors.New("cound not execute config command")
	}

	return nil
}

func errUnknownConfigKey(key string) error {
	return fmt.Errorf("unknown configuration key: %q", key)
}
