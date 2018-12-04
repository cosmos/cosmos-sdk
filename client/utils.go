package client

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PrepareCLIMainCommand adds necessary things to main gaiacli command
func PrepareCLIMainCmd(cmd *cobra.Command, envPrefix, defaultHome string) *cobra.Command {
	initConfig(cmd)
	cobra.OnInitialize(func() { initEnv(envPrefix) })
	cmd.PersistentFlags().StringP(FlagOutput, "o", "text", "Output format (text|json)")
	cmd.PersistentFlags().String(FlagHome, defaultHome, "directory for config and data")
	cmd.PersistentFlags().Bool(FlagTrace, false, "print out full stack trace on errors")
	cmd.PersistentPreRunE = concatCobraCmdFuncs(bindFlagsLoadViper, validateOutput, cmd.PersistentPreRunE)
	return cmd
}

// PrepareDMainCommand adds necessary things to main gaiad command
func PrepareDMainCmd(cmd *cobra.Command, envPrefix, defaultHome string) *cobra.Command {
	initConfig(cmd)
	cobra.OnInitialize(func() { initEnv(envPrefix) })
	cmd.PersistentFlags().String(FlagHome, defaultHome, "directory for config and data")
	cmd.PersistentFlags().Bool(FlagTrace, false, "print out full stack trace on errors")
	cmd.PersistentPreRunE = concatCobraCmdFuncs(bindFlagsLoadViper, cmd.PersistentPreRunE)
	return cmd
}

func initConfig(cmd *cobra.Command) error {
	home, err := cmd.PersistentFlags().GetString(FlagHome)
	if err != nil {
		return err
	}

	cfgFile := path.Join(home, "config", "config.toml")
	if _, err := os.Stat(cfgFile); err == nil {
		viper.SetConfigFile(cfgFile)

		if err := viper.ReadInConfig(); err != nil {
			return err
		}
	}
	return viper.BindPFlag(FlagOutput, cmd.PersistentFlags().Lookup(FlagOutput))
}

// initEnv sets prefix to use for ENV variables
func initEnv(prefix string) {
	prefix = strings.ToUpper(prefix)
	ps := prefix + "_"
	for _, e := range os.Environ() {
		kv := strings.SplitN(e, "=", 2)
		if len(kv) == 2 {
			k, v := kv[0], kv[1]
			if strings.HasPrefix(k, prefix) && !strings.HasPrefix(k, ps) {
				k2 := strings.Replace(k, prefix, ps, 1)
				os.Setenv(k2, v)
			}
		}
	}

	// env variables with prefix (eg. {{ .prefix }}_ROOT)
	viper.SetEnvPrefix(prefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()
}

// cobraCmdFuncs allows for multiple PersistentPreRun hooks to run for each command
type cobraCmdFunc func(cmd *cobra.Command, args []string) error

// Returns a single function that calls each argument function in sequence
// RunE, PreRunE, PersistentPreRunE, etc. all have this same signature
func concatCobraCmdFuncs(fs ...cobraCmdFunc) cobraCmdFunc {
	return func(cmd *cobra.Command, args []string) error {
		for _, f := range fs {
			if f != nil {
				if err := f(cmd, args); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

func validateOutput(cmd *cobra.Command, args []string) error {
	// validate output format
	output := viper.GetString(FlagOutput)
	switch output {
	case "text", "json":
	default:
		return errors.Errorf("Unsupported output format: %s", output)
	}
	return nil
}

// Bind all flags and read the config into viper
func bindFlagsLoadViper(cmd *cobra.Command, args []string) error {
	// cmd.Flags() includes flags from this command and all persistent flags from the parent
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	homeDir := viper.GetString(FlagHome)
	viper.Set(FlagHome, homeDir)
	viper.SetConfigName("config")                         // name of config file (without extension)
	viper.AddConfigPath(homeDir)                          // search root directory
	viper.AddConfigPath(filepath.Join(homeDir, "config")) // search root directory /config

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		// stderr, so if we redirect output to json file, this doesn't appear
		// fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	} else if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
		// ignore not found error, return other errors
		return err
	}
	return nil
}

// GetSDKConfig sets the address prefixes to the SDK style
func SetSDKConfig() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(sdk.Bech32PrefixAccAddr, sdk.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(sdk.Bech32PrefixValAddr, sdk.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(sdk.Bech32PrefixConsAddr, sdk.Bech32PrefixConsPub)
	config.Seal()
}
