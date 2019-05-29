package keys

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto/keys/mintkey"
	"github.com/spf13/cobra"
)

func exportKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export <name>",
		Short: "Export private keys",
		Long:  `Export a private key from the local keybase in ASCII armor format.`,
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
	encryptPassword, err := input.GetPassword("Enter passphrase to decrypt your key:", buf)
	if err != nil {
		return err
	}

	priv, err := kb.ExportPrivateKeyObject(args[0], encryptPassword)
	if err != nil {
		return err
	}

	armored := mintkey.EncryptArmorPrivKey(priv, encryptPassword)
	fmt.Println(armored)

	return nil
}
