package keys

import (
	"bufio"
	"io/ioutil"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
)

func importKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import <name> <keyfile>",
		Short: "Import private keys into the local keybase",
		Long:  "Import a ASCII armored private key into the local keybase.",
		Args:  cobra.ExactArgs(2),
		RunE:  runImportCmd,
	}
	cmd.Flags().Bool(flags.FlagLegacyKeybase, false, "Use legacy on-disk keybase")
	return cmd
}

func runImportCmd(cmd *cobra.Command, args []string) error {
	var err error
	var kb keys.Keybase
	var decryptPassword string
	if !viper.GetBool(flags.FlagLegacyKeybase) {
		kb = NewKeyring(bufio.NewReader(cmd.InOrStdin()))
		decryptPassword = DefaultKeyPass
	} else {
		cmd.PrintErrln(deprecatedKeybaseWarning)
		kb, err = NewKeyBaseFromHomeFlag()
		if err != nil {
			return err
		}
	}

	decryptPassword, err = input.GetPassword("Enter passphrase to decrypt your key:", bufio.NewReader(cmd.InOrStdin()))
	if err != nil {
		return err
	}
	bz, err := ioutil.ReadFile(args[1])
	if err != nil {
		return err
	}

	return kb.ImportPrivKey(args[0], string(bz), decryptPassword)
}
