package keys

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/input"
)

func exportKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export <name>",
		Short: "Export private keys",
		Long:  `Export a private key from the local keybase in ASCII-armored encrypted format.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runExportCmd,
	}
	return cmd
}

func runExportCmd(_ *cobra.Command, args []string) error {
	kb, err := NewKeyBaseFromHomeFlag()
	if err != nil {
		return err
	}

	buf := input.BufferStdin()
	decryptPassword, err := input.GetPassword("Enter passphrase to decrypt your key:", buf)
	if err != nil {
		return err
	}
	encryptPassword, err := input.GetPassword("Enter passphrase to encrypt the exported key:", buf)
	if err != nil {
		return err
	}

	armored, err := kb.ExportPrivKey(args[0], decryptPassword, encryptPassword)
	if err != nil {
		return err
	}

	fmt.Println(armored)
	return nil
}
