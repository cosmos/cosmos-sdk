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

func updateKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Change the password used to protect private key",
		RunE:  runUpdateCmd,
		Args:  cobra.ExactArgs(1),
	}
	cmd.Flags().Bool(flags.FlagSecretStore, false, "Use legacy secret store")
	return cmd
}

func runUpdateCmd(cmd *cobra.Command, args []string) error {
	var kb keys.Keybase

	name := args[0]
	buf := bufio.NewReader(cmd.InOrStdin())

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
