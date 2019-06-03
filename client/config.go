package client

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"

	toml "github.com/pelletier/go-toml"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/flags"
)

const (
	flagGet = "get"
)

var configDefaults = map[string]string{
	"chain-id":       "",
	"output":         "text",
	"node":           "tcp://localhost:26657",
	"broadcast-mode": "sync",
}

// ConfigCmd returns a CLI command to interactively create an application CLI
// config file.
func ConfigCmd(defaultCLIHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config <key> [value]",
		Short: "Create or query an application CLI configuration file",
		RunE:  runConfigCmd,
		Args:  cobra.RangeArgs(0, 2),
	}

	cmd.Flags().String(flags.FlagHome, defaultCLIHome,
		"set client's home directory for configuration")
	cmd.Flags().Bool(flagGet, false,
		"print configuration value or its default if unset")
	return cmd
}

func runConfigCmd(cmd *cobra.Command, args []string) error {
	cfgFile, err := ensureConfFile(viper.GetString(flags.FlagHome))
	if err != nil {
		return err
	}

	getAction := viper.GetBool(flagGet)
	if getAction && len(args) != 1 {
		return fmt.Errorf("wrong number of arguments")
	}

	// load configuration
	tree, err := loadConfigFile(cfgFile)
	if err != nil {
		return err
	}

	// print the config and exit
	if len(args) == 0 {
		s, err := tree.ToTomlString()
		if err != nil {
			return err
		}
		fmt.Print(s)
		return nil
	}

	key := args[0]

	// get config value for a given key
	if getAction {
		switch key {
		case "trace", "trust-node", "indent":
			fmt.Println(tree.GetDefault(key, false).(bool))

		default:
			if defaultValue, ok := configDefaults[key]; ok {
				fmt.Println(tree.GetDefault(key, defaultValue).(string))
				return nil
			}

			return errUnknownConfigKey(key)
		}

		return nil
	}

	if len(args) != 2 {
		return fmt.Errorf("wrong number of arguments")
	}

	value := args[1]

	// set config value for a given key
	switch key {
	case "chain-id", "output", "node", "broadcast-mode":
		tree.Set(key, value)

	case "trace", "trust-node", "indent":
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}

		tree.Set(key, boolVal)

	default:
		return errUnknownConfigKey(key)
	}

	// save configuration to disk
	if err := saveConfigFile(cfgFile, tree); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "configuration saved to %s\n", cfgFile)
	return nil
}

func ensureConfFile(rootDir string) (string, error) {
	cfgPath := path.Join(rootDir, "config")
	if err := os.MkdirAll(cfgPath, os.ModePerm); err != nil {
		return "", err
	}

	return path.Join(cfgPath, "config.toml"), nil
}

func loadConfigFile(cfgFile string) (*toml.Tree, error) {
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "%s does not exist\n", cfgFile)
		return toml.Load(``)
	}

	bz, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return nil, err
	}

	tree, err := toml.LoadBytes(bz)
	if err != nil {
		return nil, err
	}

	return tree, nil
}

func saveConfigFile(cfgFile string, tree *toml.Tree) error {
	fp, err := os.OpenFile(cfgFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer fp.Close()

	_, err = tree.WriteTo(fp)
	return err
}

func errUnknownConfigKey(key string) error {
	return fmt.Errorf("unknown configuration key: %q", key)
}
