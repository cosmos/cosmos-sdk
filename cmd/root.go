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
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	data "github.com/tendermint/go-data"
	"github.com/tendermint/go-data/base58"
	keys "github.com/tendermint/go-keys"
	"github.com/tendermint/go-keys/cryptostore"
	"github.com/tendermint/go-keys/storage/filestorage"
)

var (
	rootDir string
	output  string
	keyDir  string
	manager keys.Manager
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
	RootCmd.PersistentFlags().StringP("root", "r", os.ExpandEnv("$HOME/.tlc"), "root directory for config and data")
	RootCmd.PersistentFlags().String("keydir", "keys", "Directory to store private keys (subdir of root)")
	RootCmd.PersistentFlags().StringP("output", "o", "text", "Output format (text|json)")
	RootCmd.PersistentFlags().StringP("encoding", "e", "hex", "Binary encoding (hex|b64|btc)")
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
		// stderr, so if we redirect output to json file, this doesn't appear
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	return validateFlags(cmd)
}

// validateFlags asserts all RootCmd flags are valid
func validateFlags(cmd *cobra.Command) error {
	// validate output format
	output = viper.GetString("output")
	switch output {
	case "text", "json":
	default:
		return errors.Errorf("Unsupported output format: %s", output)
	}

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

	// store the keys directory
	keyDir = viper.GetString("keydir")
	if !filepath.IsAbs(keyDir) {
		keyDir = filepath.Join(rootDir, keyDir)
	}
	// and construct the key manager
	manager = cryptostore.New(
		cryptostore.SecretBox,
		filestorage.New(keyDir),
	)

	return nil
}
