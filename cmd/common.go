package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	data "github.com/tendermint/go-data"
	"github.com/tendermint/go-data/base58"
)

/*******

TODO

This file should move into go-common or the like as a basis for all cli tools.
It is here for experimentation of re-use between go-keys and light-client.

*********/

const (
	RootFlag     = "root"
	OutputFlag   = "output"
	EncodingFlag = "encoding"
)

func PrepareMainCmd(cmd *cobra.Command, envPrefix, defautRoot string) func() {
	cobra.OnInitialize(func() { initEnv(envPrefix) })
	cmd.PersistentFlags().StringP(RootFlag, "r", defautRoot, "root directory for config and data")
	cmd.PersistentFlags().StringP(EncodingFlag, "e", "hex", "Binary encoding (hex|b64|btc)")
	cmd.PersistentFlags().StringP(OutputFlag, "o", "text", "Output format (text|json)")
	cmd.PersistentPreRunE = multiE(bindFlags, setEncoding, validateOutput, cmd.PersistentPreRunE)
	return func() { execute(cmd) }
}

// initEnv sets to use ENV variables if set.
func initEnv(prefix string) {
	// env variables with TM prefix (eg. TM_ROOT)
	viper.SetEnvPrefix(prefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}

// execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func execute(cmd *cobra.Command) {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

type wrapE func(cmd *cobra.Command, args []string) error

func multiE(fs ...wrapE) wrapE {
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

func bindFlags(cmd *cobra.Command, args []string) error {
	// cmd.Flags() includes flags from this command and all persistent flags from the parent
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	// rootDir is command line flag, env variable, or default $HOME/.tlc
	rootDir := viper.GetString("root")
	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath(rootDir)  // search root directory

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		// stderr, so if we redirect output to json file, this doesn't appear
		// fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	} else if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
		// we ignore not found error, only parse error
		// stderr, so if we redirect output to json file, this doesn't appear
		fmt.Fprintf(os.Stderr, "%#v", err)
	}
	return nil
}

// setEncoding reads the encoding flag
func setEncoding(cmd *cobra.Command, args []string) error {
	// validate and set encoding
	enc := viper.GetString("encoding")
	switch enc {
	case "hex":
		data.Encoder = data.HexEncoder
	case "b64":
		data.Encoder = data.B64Encoder
	case "btc":
		data.Encoder = base58.BTCEncoder
	default:
		return errors.Errorf("Unsupported encoding: %s", enc)
	}
	return nil
}

func validateOutput(cmd *cobra.Command, args []string) error {
	// validate output format
	output := viper.GetString(OutputFlag)
	switch output {
	case "text", "json":
	default:
		return errors.Errorf("Unsupported output format: %s", output)
	}
	return nil
}
