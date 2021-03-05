package config

import (
	"fmt"
	"io"
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

/*

Plan:

0. When the user uses the gaiacli config node command to set the node value,
 the new value is written successfully to the ~/.gaiacli/config/config.toml

1.Create $NODE_DIR/config/client.toml to hold all client-side configuration
(e.g. keyring-backend, node, chain-id etc...)+


2. Enable get functionality of client-side configuration

get examples

config keyring-backend --get //keyring-backend returns value from config file
config with no args --get //prints default config and exits
config invalidArg --get // returns error unknown conf key
config keyring-backend output --get // returns error too many arguments

config without arguments // just print config from config file


3.Enable set functionality of client side configuration

config keyring-backend test // sets keyring-backend = test - expected behavior

config keyring-backend with no arguments // outputs default
config invalidKey invalidValue


CLI flags

do i need to add all flags? from client/flags/flags.go

"keyring-dir" : "",
"ledger": false,
"height": 0,
"gas-adjustment": flags.DefaultGasAdjustment,//float
"from": "",
"account-number": 0,
"sequence": 0,
"memo": "",
"fees": "",
"gas-prices":"",
"dry-run": false,
"generate-only":false,
"offline": false,
"yes": true, //doublec heck skip-confirmation = true








*/

var ErrWrongNumberOfArgs = fmt.Errorf("wrong number of arguments")

// is this full client side configuration?
// add height = 0
var configDefaults = map[string]string{

	"chain-id":        "",
	"keyring-backend": "os",
	"output":          "text",
	"node":            "tcp://localhost:26657",
	"broadcast-mode":  "sync",
}

// ConfigCmd returns a CLI command to interactively create an application CLI
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

	// print the default config and exit
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
		case "chain-id", "keyring-backend", "output", "node", "broadcast-mode":
			fmt.Println(tree.Get(key).(string))

		default:
			// do we need to print out default value here in case key is invalid?
			if defaultValue, ok := configDefaults[key]; ok {
				fmt.Println(tree.GetDefault(key, defaultValue).(string))
				return nil
			}

			return errUnknownConfigKey(key)
		}

		return nil
	}

	// if we set value in configuration, the number of arguments must be 2
	if len(args) != 2 {
		return ErrWrongNumberOfArgs
	}

	value := args[1]

	// set config value for a given key
	switch key {
	case "chain-id", "keyring-backend", "output", "node", "broadcast-mode":
		tree.Set(key, value)

	// do we need to check if key matches one of these?
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
	if err := os.MkdirAll(cfgPath, os.ModePerm); err != nil { // config directory
		return "", err
	}

	return path.Join(cfgPath, "client.toml"), nil
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

func saveConfigFile(cfgFile string, tree io.WriterTo) error {
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
