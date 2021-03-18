package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
)

// Cmd returns a CLI command to interactively create an application CLI
// config file.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config <key> [value]",
		Short: "Create or query an application CLI configuration file",
		RunE:  runConfigCmd,
		Args:  cobra.RangeArgs(0, 2),
	}
	return cmd
}

func runConfigCmd(cmd *cobra.Command, args []string) error {

	clientCtx := client.GetClientContextFromCmd(cmd)

	configPath := filepath.Join(clientCtx.HomeDir, "config")

	if err := ensureConfigPath(configPath); err != nil {
		return fmt.Errorf("couldn't make client config: %v", err)
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
		default:
			err := errUnknownConfigKey(key)
			return fmt.Errorf("couldn't get the value for the key: %v, error:  %v", key, err)
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
		default:
			return errUnknownConfigKey(key)
		}

		configTemplate, err := initConfigTemplate()
		if err != nil {
			return fmt.Errorf("could not initiate config template: %v", err)
		}

		cliConfigFile := filepath.Join(configPath, "client.toml")
		if err := writeConfigFile(cliConfigFile, cliConfig, configTemplate); err != nil {
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
