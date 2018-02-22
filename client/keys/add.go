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

package keys

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/go-crypto/keys"
	"github.com/tendermint/tmlibs/cli"
)

const (
	flagType     = "type"
	flagRecover  = "recover"
	flagNoBackup = "no-backup"
	flagDryRun   = "dry-run"
)

func addKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Create a new key, or import from seed",
		Long: `Add a public/private key pair to the key store.
If you select --seed/-s you can recover a key from the seed
phrase, otherwise, a new key will be generated.`,
		RunE: runAddCmd,
	}
	cmd.Flags().StringP(flagType, "t", "ed25519", "Type of private key (ed25519|secp256k1|ledger)")
	cmd.Flags().Bool(flagRecover, false, "Provide seed phrase to recover existing key instead of creating")
	cmd.Flags().Bool(flagNoBackup, false, "Don't print out seed phrase (if others are watching the terminal)")
	cmd.Flags().Bool(flagDryRun, false, "Perform action, but don't add key to local keystore")
	return cmd
}

func runAddCmd(cmd *cobra.Command, args []string) error {
	var kb keys.Keybase
	var err error
	var name, pass string

	if viper.GetBool(flagDryRun) {
		// we throw this away, so don't enforce args,
		// we want to get a new random seed phrase quickly
		kb = client.MockKeyBase()
		pass = "throwing-this-key-away"
		name = "inmemorykey"
	} else {
		if len(args) != 1 || len(args[0]) == 0 {
			return errors.New("You must provide a name for the key")
		}
		name = args[0]
		kb, err = GetKeyBase()
		if err != nil {
			return err
		}
		pass, err = getCheckPassword("Enter a passphrase for your key:", "Repeat the passphrase:")
		if err != nil {
			return err
		}
	}

	if viper.GetBool(flagRecover) {
		seed, err := getPassword("Enter your recovery seed phrase:")
		if err != nil {
			return err
		}
		info, err := kb.Recover(name, pass, seed)
		if err != nil {
			return err
		}
		// print out results without the seed phrase
		viper.Set(flagNoBackup, true)
		printCreate(info, "")
	} else {
		algo := keys.CryptoAlgo(viper.GetString(flagType))
		info, seed, err := kb.Create(name, pass, algo)
		if err != nil {
			return err
		}
		printCreate(info, seed)
	}
	return nil
}

// addOutput lets us json format the data
type addOutput struct {
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
		out := addOutput{Key: info}
		if !viper.GetBool(flagNoBackup) {
			out.Seed = seed
		}
		json, err := MarshalJSON(out)
		if err != nil {
			panic(err) // really shouldn't happen...
		}
		fmt.Println(string(json))
	}
}
