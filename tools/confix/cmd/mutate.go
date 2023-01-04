package cmd

import (
	"github.com/spf13/cobra"
)

// UpdateCommand returns a CLI command to interactively update an application config value.
func UpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <config> <key> <value>",
		Short: "Update an application config value",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return cmd
}

// GetCommand returns a CLI command to interactively get an application config value.
func GetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <config> <key>",
		Short: "Get an application config value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return cmd
}

// func runConfigCmd(cmd *cobra.Command, args []string) error {
// 	clientCtx := client.GetClientContextFromCmd(cmd)
// 	configPath := filepath.Join(clientCtx.HomeDir, "config")

// 	conf, err := getClientConfig(configPath, clientCtx.Viper)
// 	if err != nil {
// 		return fmt.Errorf("couldn't get client config: %v", err)
// 	}

// 	switch len(args) {
// 	case 0:
// 		// print all client config fields to stdout
// 		s, err := json.MarshalIndent(conf, "", "\t")
// 		if err != nil {
// 			return err
// 		}
// 		cmd.Println(string(s))

// 	case 1:
// 		// it's a get
// 		key := args[0]

// 		switch key {
// 		case flags.FlagChainID:
// 			cmd.Println(conf.ChainID)
// 		case flags.FlagKeyringBackend:
// 			cmd.Println(conf.KeyringBackend)
// 		case flags.FlagOutput:
// 			cmd.Println(conf.Output)
// 		case flags.FlagNode:
// 			cmd.Println(conf.Node)
// 		case flags.FlagBroadcastMode:
// 			cmd.Println(conf.BroadcastMode)
// 		default:
// 			err := errUnknownConfigKey(key)
// 			return fmt.Errorf("couldn't get the value for the key: %v, error:  %v", key, err)
// 		}

// 	case 2:
// 		// it's set
// 		key, value := args[0], args[1]

// 		switch key {
// 		case flags.FlagChainID:
// 			conf.SetChainID(value)
// 		case flags.FlagKeyringBackend:
// 			conf.SetKeyringBackend(value)
// 		case flags.FlagOutput:
// 			conf.SetOutput(value)
// 		case flags.FlagNode:
// 			conf.SetNode(value)
// 		case flags.FlagBroadcastMode:
// 			conf.SetBroadcastMode(value)
// 		default:
// 			return errUnknownConfigKey(key)
// 		}

// 		confFile := filepath.Join(configPath, "client.toml")
// 		if err := writeConfigToFile(confFile, conf); err != nil {
// 			return fmt.Errorf("could not write client config to the file: %v", err)
// 		}

// 	default:
// 		panic("cound not execute config command")
// 	}

// 	return nil
// }

// func errUnknownConfigKey(key string) error {
// 	return fmt.Errorf("unknown configuration key: %q", key)
// }
