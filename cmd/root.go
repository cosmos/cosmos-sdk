// Copyright © 2017 Ethan Frey
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	rootDir string
	format  string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "keys",
	Short: "Key manager for tendermint clients",
	Long: `Keys allows you to manage your local keystore for tendermint.

These keys may be in any format supported by go-crypto and can be
used by light-clients, full nodes, or any other application that
needs to sign with a private key.`,
	PersistentPreRunE: bindFlags,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initEnv)
	RootCmd.PersistentFlags().StringP("root", "r", os.ExpandEnv("$HOME/.tlc"), "root directory for config and data (default is TM_ROOT or $HOME/.tlc)")
	RootCmd.PersistentFlags().StringP("format", "f", "text", "Output format (text|json)")
}

// initEnv sets to use ENV variables if set.
func initEnv() {
	// env variables with TM prefix (eg. TM_ROOT)
	viper.SetEnvPrefix("TM")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}

func bindFlags(cmd *cobra.Command, args []string) error {
	// cmd.Flags() includes flags from this command and all persistent flags from the parent
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	// rootDir is command line flag, env variable, or default $HOME/.tlc
	rootDir = viper.GetString("root")
	viper.SetConfigName("keys")  // name of config file (without extension)
	viper.AddConfigPath(rootDir) // search root directory

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	return validateFlags(cmd)
}

// validateFlags asserts all RootCmd flags are valid
func validateFlags(cmd *cobra.Command) error {
	format = viper.GetString("format")
	switch format {
	case "text", "json":
		return nil
	default:
		return errors.Errorf("Unsupported format: %s", format)
	}
}
