package keys

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/flags"
)

const flagListNames = "list-names"

// ListKeysCmd lists all keys in the key store.
func ListKeysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all keys",
		Long: `Return a list of all public keys stored by this key manager
along with their associated name and address.`,
		RunE: runListCmd,
	}
	cmd.Flags().Bool(flags.FlagIndentResponse, false, "Add indent to JSON response")
	cmd.Flags().BoolP(flagListNames, "n", false, "List names only")
	return cmd
}

func runListCmd(cmd *cobra.Command, _ []string) error {
	kb, err := NewKeyringFromHomeFlag(cmd.InOrStdin())
	if err != nil {
		return err
	}

	infos, err := kb.List()
	if err != nil {
		return err
	}

	if !viper.GetBool(flagListNames) {
		printInfos(infos)
		return nil
	}

	cmd.SetOut(cmd.OutOrStdout())
	for _, info := range infos {
		cmd.Println(info.GetName())
	}

	return nil
}
