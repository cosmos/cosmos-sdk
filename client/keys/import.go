package keys

import (
	"bufio"
	"io/ioutil"

	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/spf13/cobra"
)

func importKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import <name> <keyfile>",
		Short: "Import private keys into the local keybase",
		Long:  "Import a ASCII armored private key into the local keybase.",
		Args:  cobra.ExactArgs(2),
		RunE:  runImportCmd,
	}
	return cmd
}

func runImportCmd(cmd *cobra.Command, args []string) error {
	kb, err := NewKeyBaseFromHomeFlag()
	if err != nil {
		return err
	}

	bz, err := ioutil.ReadFile(args[1])
	if err != nil {
		return err
	}

	buf := bufio.NewReader(cmd.InOrStdin())
	passphrase, err := input.GetPassword("Enter passphrase to decrypt your key:", buf)
	if err != nil {
		return err
	}

	return kb.ImportPrivKey(args[0], string(bz), passphrase)
}
