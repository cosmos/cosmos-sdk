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
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	keys "github.com/tendermint/go-keys"
	"github.com/tendermint/go-keys/cryptostore"
	"github.com/tendermint/go-keys/storage/filestorage"
)

const KeySubdir = "keys"

var (
	Manager keys.Manager
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "keys",
	Short: "Key manager for tendermint clients",
	Long: `Keys allows you to manage your local keystore for tendermint.

These keys may be in any format supported by go-crypto and can be
used by light-clients, full nodes, or any other application that
needs to sign with a private key.`,
}

// SetupKeys must be registered in main() on the top level command
// here this is RootCmd, but if we embed keys in eg. light-client,
// that must be responsible for the update
func SetupKeys(cmd *cobra.Command, args []string) error {
	// store the keys directory
	rootDir := viper.GetString("root")
	keyDir := filepath.Join(rootDir, KeySubdir)
	// and construct the key manager
	Manager = cryptostore.New(
		cryptostore.SecretBox,
		filestorage.New(keyDir),
	)
	return nil
}
