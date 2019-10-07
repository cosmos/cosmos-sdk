package keys

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
)

func listKeysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all keys",
		Long: `Return a list of all public keys stored by this key manager
along with their associated name and address.`,
		RunE: runListCmd,
	}
	cmd.Flags().Bool(flags.FlagIndentResponse, false, "Add indent to JSON response")
	cmd.Flags().Bool(flags.FlagLegacyKeybase, false, "Use legacy on-disk keybase")
	return cmd
}

func runListCmd(cmd *cobra.Command, args []string) error {
	var kb keys.Keybase
	if !viper.GetBool(flags.FlagLegacyKeybase) {
		kb = NewKeyring(cmd.InOrStdin())
	} else {
		cmd.PrintErrln(DeprecatedKeybaseWarning)
		var err error
		kb, err = NewKeyBaseFromHomeFlag()
		if err != nil {
			return err
		}
	}

	infos, err := kb.List()
	if err == nil {
		printInfos(infos)
	}
	return err
}
