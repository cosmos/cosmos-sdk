package keys

import (
	"bufio"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
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

	kb, err := keyring.New(sdk.KeyringServiceName(), viper.GetString(flags.FlagKeyringBackend), viper.GetString(flags.FlagHome), buf)
	if err != nil {
		return err
	}

	for _, name := range args {
		info, err := kb.Key(name)
		if err != nil {
			return err
		}

		// confirm deletion, unless -y is passed
		if !viper.GetBool(flagYes) {
			if yes, err := input.GetConfirmation("Key reference will be deleted. Continue?", buf, cmd.ErrOrStderr()); err != nil {
				return err
			} else if !yes {
				continue
			}
		}

		if err := kb.Delete(name); err != nil {
			return err
		}

		if info.GetType() == keyring.TypeLedger || info.GetType() == keyring.TypeOffline {
			cmd.PrintErrln("Public key reference deleted")
			continue
		}
		cmd.PrintErrln("Key deleted forever (uh oh!)")
	}

	return nil
}
