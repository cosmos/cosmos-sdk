package keys

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
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

	cmd.Flags().BoolP(flagListNames, "n", false, "List names only")
	return cmd
}

func runListCmd(cmd *cobra.Command, _ []string) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}

	infos, err := clientCtx.Keyring.List()
	if err != nil {
		return err
	}

	if ok, _ := cmd.Flags().GetBool(flagListNames); !ok {
		printInfos(cmd.OutOrStdout(), infos, clientCtx.OutputFormat)
		return nil
	}

	for _, info := range infos {
		cmd.Println(info.GetName())
	}

	return nil
}
