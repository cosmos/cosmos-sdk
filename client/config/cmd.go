package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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

	conf, err := getClientConfig(configPath, clientCtx.Viper)
	if err != nil {
		return fmt.Errorf("couldn't get client config: %v", err)
	}

	switch len(args) {
	case 0:
		// print all client config fields to stdout
		s, err := json.MarshalIndent(conf, "", "\t")
		if err != nil {
			return err
		}
		cmd.Println("showing configuration at", clientCtx.HomeDir)
		cmd.Println(string(s))

	case 1:
		// it's a get
		key := args[0]

		cmd.Println("showing configuration at", clientCtx.HomeDir)
		switch key {
		case flags.FlagChainID:
			cmd.Println(conf.ChainID)
		case flags.FlagKeyringBackend:
			cmd.Println(conf.KeyringBackend)
		case flags.FlagOutput:
			cmd.Println(conf.Output)
		case flags.FlagNode:
			cmd.Println(conf.Node)
		case flags.FlagBroadcastMode:
			cmd.Println(conf.BroadcastMode)
		default:
			err := errUnknownConfigKey(key)
			return fmt.Errorf("couldn't get the value for the key: %v, error:  %v", key, err)
		}

	case 2:
		// it's set
		key, value := args[0], args[1]

		switch key {
		case flags.FlagChainID:
			conf.SetChainID(value)
		case flags.FlagKeyringBackend:
			conf.SetKeyringBackend(value)
		case flags.FlagOutput:
			conf.SetOutput(value)
		case flags.FlagNode:
			conf.SetNode(value)
		case flags.FlagBroadcastMode:
			conf.SetBroadcastMode(value)
		case flags.FlagHome:
			// check if passed value is a valid path for a home directory and create if it is.
			if _, err = os.Stat(value); err != nil {
				if err = ensureConfigPath(value); err != nil {
					return fmt.Errorf("could not create config folder: %s", value)
				}
			}

			// TODO: Change to TOML
			// TODO: Change default node home
			//homeFilePath := filepath.Join(simapp.DefaultNodeHome, "config", "home.txt")
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("could not query user home directory: %v", err)
			}
			homeFilePath := filepath.Join(homeDir, ".simapp", "config", "home.txt")
			fmt.Println("selected node configuration at", value)
			err = writeHomeDirToFile(homeFilePath, value)
			if err != nil {
				return fmt.Errorf("could not write new home directory to the configuration file at %s", homeFilePath)
			}

			return nil

		default:
			return errUnknownConfigKey(key)
		}

		confFile := filepath.Join(configPath, "client.toml")
		if err := writeConfigToFile(confFile, conf); err != nil {
			return fmt.Errorf("could not write client config to the file: %v", err)
		}

	default:
		panic("cound not execute config command")
	}

	return nil
}

func errUnknownConfigKey(key string) error {
	return fmt.Errorf("unknown configuration key: %q", key)
}
