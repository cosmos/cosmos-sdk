package keys

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/flags"
)

// Commands registers a sub-tree of commands to interact with
// local private key storage.
func Commands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "Add or view local private keys",
		Long: `Keys allows you to manage your local keystore for tendermint.

    These keys may be in any format supported by go-crypto and can be
    used by light-clients, full nodes, or any other application that
    needs to sign with a private key.`,
	}
	cmd.AddCommand(
		mnemonicKeyCommand(),
		addKeyCommand(),
		exportKeyCommand(),
		importKeyCommand(),
		listKeysCmd(),
		showKeysCmd(),
		flags.LineBreak,
		deleteKeyCommand(),
		updateKeyCommand(),
		parseKeyStringCommand(),
	)
	return cmd
}
