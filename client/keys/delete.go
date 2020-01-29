package keys

import (
	"bufio"
	"errors"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	flagYes   = "yes"
	flagForce = "force"
)

// DeleteKeyCommand deletes a key from the key store.
func DeleteKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name>...",
		Short: "Delete the given keys",
		Long: `Delete keys from the Keybase backend.

Note that removing offline or ledger keys will remove
only the public key references stored locally, i.e.
private keys stored in a ledger device cannot be deleted with the CLI.
`,
		RunE: runDeleteCmd,
		Args: cobra.MinimumNArgs(1),
	}

	cmd.Flags().BoolP(flagYes, "y", false,
		"Skip confirmation prompt when deleting offline or ledger key references")
	cmd.Flags().BoolP(flagForce, "f", false,
		"Remove the key unconditionally without asking for the passphrase. Deprecated.")
	return cmd
}

func runDeleteCmd(cmd *cobra.Command, args []string) error {
	buf := bufio.NewReader(cmd.InOrStdin())

	kb, err := keys.NewKeyring(sdk.KeyringServiceName(), viper.GetString(flags.FlagKeyringBackend), viper.GetString(flags.FlagHome), buf)
	if err != nil {
		return err
	}

	for _, name := range args {
		info, err := kb.Get(name)
		if err != nil {
			return err
		}

		if info.GetType() == keys.TypeLedger || info.GetType() == keys.TypeOffline {
			// confirm deletion, unless -y is passed
			if !viper.GetBool(flagYes) {
				if err := confirmDeletion(buf); err != nil {
					return err
				}
			}

			if err := kb.Delete(name, "", true); err != nil {
				return err
			}
			cmd.PrintErrln("Public key reference deleted")
			return nil
		}

		// old password and skip flag arguments are ignored
		if err := kb.Delete(name, "", true); err != nil {
			return err
		}
		cmd.PrintErrln("Key deleted forever (uh oh!)")
	}

	return nil
}

func confirmDeletion(buf *bufio.Reader) error {
	answer, err := input.GetConfirmation("Key reference will be deleted. Continue?", buf)
	if err != nil {
		return err
	}
	if !answer {
		return errors.New("aborted")
	}
	return nil
}
