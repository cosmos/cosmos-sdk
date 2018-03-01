package keys

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
)

func deleteKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete the given key",
		RunE:  runDeleteCmd,
	}
	return cmd
}

func runDeleteCmd(cmd *cobra.Command, args []string) error {
	if len(args) != 1 || len(args[0]) == 0 {
		return errors.New("You must provide a name for the key")
	}
	name := args[0]

	buf := client.BufferStdin()
	oldpass, err := client.GetPassword(
		"DANGER - enter password to permanently delete key:", buf)
	if err != nil {
		return err
	}

	kb, err := GetKeyBase()
	if err != nil {
		return err
	}

	err = kb.Delete(name, oldpass)
	if err != nil {
		return err
	}
	fmt.Println("Password deleted forever (uh oh!)")
	return nil
}
