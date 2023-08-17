package keys

import (
	"bufio"
	"os"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/input"
)

// ImportKeyCommand imports private keys from a keyfile.
func ImportKeyCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "import <name> <keyfile>",
		Short: "Import private keys into the local keybase",
		Long:  "Import a ASCII armored private key into the local keybase.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			buf := bufio.NewReader(clientCtx.Input)

			bz, err := os.ReadFile(args[1])
			if err != nil {
				return err
			}

			passphrase, err := input.GetPassword("Enter passphrase to decrypt your key:", buf)
			if err != nil {
				return err
			}

			return clientCtx.Keyring.ImportPrivKey(args[0], string(bz), passphrase)
		},
	}
}

func ImportKeyHexCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "import-hex <name> <hex> <key-type>",
		Short: "Import private keys into the local keybase",
		Long:  "Import hex encoded private key into the local keybase.",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			return clientCtx.Keyring.ImportPrivKeyHex(args[0], args[1], args[2])
		},
	}
}
