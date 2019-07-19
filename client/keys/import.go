package keys

import (
	"bufio"
	"fmt"
	"io/ioutil"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func importKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import <name> <keyfile>",
		Short: "Import private keys into the local keybase",
		Long:  "Import a ASCII armored private key into the local keybase.",
		Args:  cobra.ExactArgs(2),
		RunE:  runImportCmd,
	}
	cmd.Flags().Bool(flags.FlagSecretStore, false, "Use legacy secret store")
	return cmd
}

func runImportCmd(cmd *cobra.Command, args []string) error {
	var kb keys.Keybase

	if viper.GetBool(flags.FlagSecretStore) {
		fmt.Println("Using deprecated secret store. This will be removed in a future release.")
		var err error
		kb, err = NewKeyBaseFromHomeFlag()
		if err != nil {
			return err
		}
	} else {
		kb = NewKeyringKeybase(cmd.InOrStdin())

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
