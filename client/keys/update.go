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

	buf := client.BufferStdin()
	oldpass, err := client.GetPassword(
		"Enter the current passphrase:", buf)
	if err != nil {
		return err
	}
	newpass, err := client.GetCheckPassword(
		"Enter the new passphrase:",
		"Repeat the new passphrase:", buf)
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
