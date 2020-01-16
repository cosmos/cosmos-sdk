package keys

import (
	"bufio"
	"io/ioutil"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/input"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ImportKeyCommand imports private keys from a keyfile.
func ImportKeyCommand(config *sdk.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "import <name> <keyfile>",
		Short: "Import private keys into the local keybase",
		Long:  "Import a ASCII armored private key into the local keybase.",
		Args:  cobra.ExactArgs(2),
		RunE:  runImportCmd(config),
	}
}

func runImportCmd(config *sdk.Config) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		buf := bufio.NewReader(cmd.InOrStdin())
		kb, err := NewKeyringFromHomeFlag(buf, config)
		if err != nil {
			return err
		}

		bz, err := ioutil.ReadFile(args[1])
		if err != nil {
			return err
		}

		passphrase, err := input.GetPassword("Enter passphrase to decrypt your key:", buf)
		if err != nil {
			return err
		}

		return kb.ImportPrivKey(args[0], string(bz), passphrase)
	}
}
