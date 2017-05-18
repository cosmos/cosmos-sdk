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
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "Create a new public/private key pair",
	Long: `Add a public/private key pair to the key store.
The password muts be entered in the terminal and not
passed as a command line argument for security.`,
	RunE: newPassword,
}

func init() {
	RootCmd.AddCommand(newCmd)
	newCmd.Flags().StringP("type", "t", "ed25519", "Type of key (ed25519|secp256k1)")
}

func newPassword(cmd *cobra.Command, args []string) error {
	if len(args) != 1 || len(args[0]) == 0 {
		return errors.New("You must provide a name for the key")
	}
	name := args[0]
	algo := viper.GetString("type")

	pass, err := getCheckPassword("Enter a passphrase:", "Repeat the passphrase:")
	if err != nil {
		return err
	}

	info, err := GetKeyManager().Create(name, pass, algo)
	if err == nil {
		printInfo(info)
	}
	return err
}
