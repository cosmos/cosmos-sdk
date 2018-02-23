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
)

func updateKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Change the password used to protect private key",
		RunE:  runUpdateCmd,
	}
	return cmd
}

func runUpdateCmd(cmd *cobra.Command, args []string) error {
	if len(args) != 1 || len(args[0]) == 0 {
		return errors.New("You must provide a name for the key")
	}
	name := args[0]

	oldpass, err := client.GetPassword("Enter the current passphrase:")
	if err != nil {
		return err
	}
	newpass, err := client.GetCheckPassword("Enter the new passphrase:", "Repeat the new passphrase:")
	if err != nil {
		return err
	}

	kb, err := GetKeyBase()
	if err != nil {
		return err
	}
	err = kb.Update(name, oldpass, newpass)
	if err != nil {
		return err
	}
	fmt.Println("Password successfully updated!")
	return nil
}
