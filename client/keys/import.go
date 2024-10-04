package keys

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/version"
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
			name := args[0]
			if strings.TrimSpace(name) == "" {
				return errors.New("the provided name is invalid or empty after trimming whitespace")
			}
			buf := bufio.NewReader(clientCtx.Input)

			armor, err := os.ReadFile(args[1])
			if err != nil {
				return err
			}

			passphrase, err := input.GetPassword("Enter passphrase to decrypt your key:", buf)
			if err != nil {
				return err
			}

			return clientCtx.Keyring.ImportPrivKey(name, string(armor), passphrase)
		},
	}
}

func ImportKeyHexCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import-hex <name> [hex]",
		Short: "Import private keys into the local keybase",
		Long:  fmt.Sprintf("Import hex encoded private key into the local keybase.\nSupported key-types can be obtained with:\n%s list-key-types", version.AppName),
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			name := args[0]
			if strings.TrimSpace(name) == "" {
				return errors.New("the provided name is invalid or empty after trimming whitespace")
			}
			keyType, _ := cmd.Flags().GetString(flags.FlagKeyType)
			var hexKey string
			if len(args) == 2 {
				hexKey = args[1]
			} else {
				buf := bufio.NewReader(clientCtx.Input)
				hexKey, err = input.GetPassword("Enter hex private key:", buf)
				if err != nil {
					return err
				}
			}
			return clientCtx.Keyring.ImportPrivKeyHex(name, hexKey, keyType)
		},
	}
	cmd.Flags().String(flags.FlagKeyType, string(hd.Secp256k1Type), "private key signing algorithm kind")
	return cmd
}
