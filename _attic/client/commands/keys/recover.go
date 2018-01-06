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
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
)

// recoverCmd represents the recover command
var recoverCmd = &cobra.Command{
	Use:   "recover [name]",
	Short: "Recover a private key from a seed phrase",
	Long: `Recover a private key from a seed phrase.

I really hope you wrote this down when you created the new key.
The seed is only displayed on creation, never again.

You can also use this to copy a key between multiple testnets,
simply by "recovering" the key in the other nets you want to copy
to.  Of course, it has no coins on the other nets, just the same address.`,
	RunE: runRecoverCmd,
}

func runRecoverCmd(cmd *cobra.Command, args []string) error {
	if len(args) != 1 || len(args[0]) == 0 {
		return errors.New("You must provide a name for the key")
	}
	name := args[0]

	pass, err := getPassword("Enter the new passphrase:")
	if err != nil {
		return err
	}

	// not really a password... huh?
	seed, err := getSeed("Enter your recovery seed phrase:")
	if err != nil {
		return err
	}

	info, err := GetKeyManager().Recover(name, pass, seed)
	if err != nil {
		return err
	}
	printInfo(info)
	return nil
}
