package keys

import (
	"bufio"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
)

// UpdateKeyCommand changes the password of a key in the keybase.
// It takes no effect on keys managed by new the keyring-based keybase implementation.
func UpdateKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Change the password used to protect private key",
		Deprecated: `it takes no effect with the new keyring
based backend and is provided only for backward compatibility with the
legacy LevelDB based backend.
Refer to your operating system's manual to learn how to change your
keyring's password.
`,
		RunE: runUpdateCmd,
		Args: cobra.ExactArgs(1),
	}
	return cmd
}

func runUpdateCmd(cmd *cobra.Command, args []string) error {
	name := args[0]

	buf := bufio.NewReader(cmd.InOrStdin())
	kb, err := NewKeyBaseFromDir(viper.GetString(flags.FlagHome))
	if err != nil {
		return err
	}
	oldpass, err := input.GetPassword("Enter the current passphrase:", buf)
	if err != nil {
		return err
	}

	getNewpass := func() (string, error) {
		return input.GetCheckPassword(
			"Enter the new passphrase:",
			"Repeat the new passphrase:", buf)
	}
	if err := kb.Update(name, oldpass, getNewpass); err != nil {
		return err
	}

	cmd.PrintErrln("Password successfully updated!")
	return nil
}
