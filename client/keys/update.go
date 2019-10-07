package keys

import (
	"bufio"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/input"
)

func updateKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Change the password used to protect private key",
		Long: `Update private key's encryption passphrase. This command is no-op
when keys are stored in the operating system credential storage.`,
		RunE: runUpdateCmd,
		Args: cobra.ExactArgs(1),
	}
	return cmd
}

func runUpdateCmd(cmd *cobra.Command, args []string) error {
	name := args[0]

	buf := bufio.NewReader(cmd.InOrStdin())
	cmd.PrintErrln(DeprecatedKeybaseWarning)
	kb, err := NewKeyBaseFromHomeFlag()
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
