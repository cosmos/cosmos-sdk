package keys

import (
	"bufio"
	"io/ioutil"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
)

// ImportKeyCommand imports private keys from a keyfile.
func ImportKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import <name> <keyfile>",
		Short: "Import private keys into the local keybase",
		Long:  "Import a ASCII armored private key into the local keybase.",
		Args:  cobra.ExactArgs(2),
		RunE:  runImportCmd,
	}
	cmd.Flags().String(flagKeyAlgo, string(keys.Secp256k1), "Key signing algorithm to import key for")
	return cmd
}

func runImportCmd(cmd *cobra.Command, args []string) error {
	buf := bufio.NewReader(cmd.InOrStdin())
	kb, err := NewKeyringFromHomeFlag(buf)
	if err != nil {
		return err
	}

	algo := keys.SigningAlgo(viper.GetString(flagKeyAlgo))
	if !keys.IsAlgoSupported(algo, kb.SupportedAlgos()) {
		return keys.ErrUnsupportedSigningAlgo
	}

	bz, err := ioutil.ReadFile(args[1])
	if err != nil {
		return err
	}

	passphrase, err := input.GetPassword("Enter passphrase to decrypt your key:", buf)
	if err != nil {
		return err
	}

	return kb.ImportPrivKey(args[0], string(bz), passphrase, algo)
}
