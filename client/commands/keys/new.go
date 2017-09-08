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

	"github.com/pkg/errors"
	"github.com/tendermint/go-crypto/keys"
	"github.com/tendermint/go-wire/data"
	"github.com/tendermint/tmlibs/cli"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	flagType     = "type"
	flagNoBackup = "no-backup"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new [name]",
	Short: "Create a new public/private key pair",
	Long: `Add a public/private key pair to the key store.
The password muts be entered in the terminal and not
passed as a command line argument for security.`,
	RunE: runNewCmd,
}

func init() {
	newCmd.Flags().StringP(flagType, "t", "ed25519", "Type of key (ed25519|secp256k1)")
	newCmd.Flags().Bool(flagNoBackup, false, "Don't print out seed phrase (if others are watching the terminal)")
}

func runNewCmd(cmd *cobra.Command, args []string) error {
	if len(args) != 1 || len(args[0]) == 0 {
		return errors.New("You must provide a name for the key")
	}
	name := args[0]
	algo := viper.GetString(flagType)

	pass, err := getCheckPassword("Enter a passphrase:", "Repeat the passphrase:")
	if err != nil {
		return err
	}

	info, seed, err := GetKeyManager().Create(name, pass, algo)
	if err == nil {
		printCreate(info, seed)
	}
	return err
}

type NewOutput struct {
	Key  keys.Info `json:"key"`
	Seed string    `json:"seed"`
}

func printCreate(info keys.Info, seed string) {
	switch viper.Get(cli.OutputFlag) {
	case "text":
		printInfo(info)
		// print seed unless requested not to.
		if !viper.GetBool(flagNoBackup) {
			fmt.Println("**Important** write this seed phrase in a safe place.")
			fmt.Println("It is the only way to recover your account if you ever forget your password.\n")
			fmt.Println(seed)
		}
	case "json":
		out := NewOutput{Key: info}
		if !viper.GetBool(flagNoBackup) {
			out.Seed = seed
		}
		json, err := data.ToJSON(out)
		if err != nil {
			panic(err) // really shouldn't happen...
		}
		fmt.Println(string(json))
	}
}
