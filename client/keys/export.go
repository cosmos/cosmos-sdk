package keys

import (
	"bufio"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/input"
)

const (
	flagHex = "hex"
)

// ExportKeyCommand exports private keys from the key store.
func ExportKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export <name>",
		Short: "Export private keys",
		Long:  `Export a private key from the local keybase in ASCII-armored encrypted format.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			buf := bufio.NewReader(cmd.InOrStdin())
			clientCtx := client.GetClientContextFromCmd(cmd)

			encryptPassword, err := input.GetPassword("Enter passphrase to encrypt the exported key:", buf)
			if err != nil {
				return err
			}

			hex, _ := cmd.Flags().GetBool(flagHex)

			if !hex {
				armored, err := clientCtx.Keyring.ExportPrivKeyArmor(args[0], encryptPassword)
				if err != nil {
					return err
				}

				cmd.Println(armored)
				return nil
			} else {
				hexPrivKey, err := clientCtx.Keyring.ExportPrivKeyHex(args[0])

				if err != nil {
					return err
				}

				cmd.Println(hexPrivKey)
				return nil
			}

		},
	}

	cmd.Flags().BoolP(flagHex, "x", false, "Export unarmored hex privkey")

	return cmd
}
