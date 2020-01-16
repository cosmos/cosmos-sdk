package keys

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Commands registers a sub-tree of commands to interact with
// local private key storage.
func Commands(config *sdk.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "Add or view local private keys",
		Long: `Keys allows you to manage your local keystore for tendermint.

    These keys may be in any format supported by go-crypto and can be
    used by light-clients, full nodes, or any other application that
    needs to sign with a private key.`,
	}
	cmd.AddCommand(
		MnemonicKeyCommand(),
		AddKeyCommand(config),
		ExportKeyCommand(config),
		ImportKeyCommand(config),
		ListKeysCmd(config),
		ShowKeysCmd(config),
		flags.LineBreak,
		DeleteKeyCommand(config),
		UpdateKeyCommand(config),
		ParseKeyStringCommand(config),
		MigrateCommand(config),
	)
	cmd.PersistentFlags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|test)")
	viper.BindPFlag(flags.FlagKeyringBackend, cmd.Flags().Lookup(flags.FlagKeyringBackend))
	return cmd
}
