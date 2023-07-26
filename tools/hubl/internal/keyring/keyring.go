package keyring

import (
	"context"
	"fmt"
	"io"
	"path"

	"github.com/spf13/cobra"

	_ "cosmossdk.io/api/cosmos/crypto/ed25519"
	_ "cosmossdk.io/api/cosmos/crypto/secp256k1"
	_ "cosmossdk.io/api/cosmos/crypto/secp256r1"
	"cosmossdk.io/tools/hubl/internal/config"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func Cmd() *cobra.Command {
	keyringCmd := &cobra.Command{
		Use:   "keys",
		Short: "Global keyring management for Hubl",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cmd.SetContext(context.Background())
			backend, err := cmd.Flags().GetString(flags.FlagKeyringBackend)
			if err != nil {
				return err
			}

			keyring, err := createKeyring(cmd.InOrStdin(), backend)
			if err != nil {
				return err
			}

			clientCtx := client.Context{}.
				WithKeyring(keyring)
				// TODO: add more config options here

			if err := client.SetCmdClientContextHandler(clientCtx, cmd); err != nil {
				return err
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	keyringCmd.AddCommand(
		keys.AddKeyCommand(),
		keys.ListKeysCmd(),
	)

	keyringCmd.PersistentFlags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|kwallet|pass|test|memory)")

	return keyringCmd
}

func ChainCmd(name string) *cobra.Command {
	keyringCmd := &cobra.Command{
		Use:   "keys",
		Short: fmt.Sprintf("Keyring management for %s chain", name),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	return keyringCmd
}

func createKeyring(input io.Reader, backend string) (keyring.Keyring, error) {
	keyringDir := path.Join(config.DefaultConfigDirName, "keyring")

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	return keyring.New(sdk.KeyringServiceName(), backend, keyringDir, input, cdc)
}
