package keys

import (
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
)

var showKeysCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show key info for the given name",
	Long:  `Return public details of one local key.`,
	RunE:  runShowCmd,
}

func runShowCmd(cmd *cobra.Command, args []string) error {
	if len(args) != 1 || len(args[0]) == 0 {
		return errors.New("You must provide a name for the key")
	}
	name := args[0]

	kb, err := GetKeyBase()
	if err != nil {
		return err
	}

	info, err := kb.Get(name)
	if err == nil {
		printInfo(info)
	}
	return err
}
