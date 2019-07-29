package keys

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	cmd.Flags().Bool(flags.FlagSecretStore, false, "Use legacy secret store")
	return cmd
}

func runListCmd(cmd *cobra.Command, args []string) error {
	var kb keys.Keybase

	if viper.GetBool(flags.FlagSecretStore) {
		os.Stderr.WriteString("Using deprecated secret store. This will be removed in a future release.")
		var err error
		kb, err = NewKeyBaseFromHomeFlag()
		if err != nil {
			return err
		}
	} else {
		kb = NewKeyringKeybase(cmd.InOrStdin())
	}

	infos, err := kb.List()
	if err == nil {
		printInfos(infos)
	}
	return err
}
