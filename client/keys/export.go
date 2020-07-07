package keys

import (
	"bufio"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ExportKeyCommand exports private keys from the key store.
func ExportKeyCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "export <name>",
		Short: "Export private keys",
		Long:  `Export a private key from the local keybase in ASCII-armored encrypted format.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			buf := bufio.NewReader(cmd.InOrStdin())

			backend, _ := cmd.Flags().GetString(flags.FlagKeyringBackend)
			homeDir, _ := cmd.Flags().GetString(flags.FlagHome)
			kb, err := keyring.New(sdk.KeyringServiceName(), backend, homeDir, buf)
			if err != nil {
				return err
			}

			encryptPassword, err := input.GetPassword("Enter passphrase to encrypt the exported key:", buf)
			if err != nil {
				return err
			}

			armored, err := kb.ExportPrivKeyArmor(args[0], encryptPassword)
			if err != nil {
				return err
			}

			cmd.Println(armored)
			return nil
		},
	}
}
