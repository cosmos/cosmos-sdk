package keys

import (
	"bufio"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
)

func exportKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export <name>",
		Short: "Export private keys",
		Long:  `Export a private key from the local keybase in ASCII-armored encrypted format.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runExportCmd,
	}
	cmd.Flags().Bool(flags.FlagLegacy, false, "Use legacy secret store")
	return cmd
}

func runExportCmd(cmd *cobra.Command, args []string) error {
	var kb keys.Keybase

	buf := bufio.NewReader(cmd.InOrStdin())

	if viper.GetBool(flags.FlagLegacy) {
		fmt.Println("IMPORTANT: using deprecated secret store. This will be removed in a future release.")
		var err error
		kb, err = NewKeyBaseFromHomeFlag()
		if err != nil {
			return err
		}
	} else {
		kb = NewKeyringKeybase(buf)
	}

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

	cmd.Println(armored)
	return nil
}
