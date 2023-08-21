package keyring

import (
	"context"
	"fmt"
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
	keys := keys.Commands()

	keyringCmd := &cobra.Command{
		Use:   "keys",
		Short: "Global keyring management for Hubl",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cmd.SetContext(context.Background())
			backend, err := cmd.Flags().GetString(flags.FlagKeyringBackend)
			if err != nil {
				return err
			}

			registry := codectypes.NewInterfaceRegistry()
			cdc := codec.NewProtoCodec(registry)

			keyringDir := path.Join(config.DefaultConfigDirName, "keyring")
			kr, err := keyring.New(sdk.KeyringServiceName(), backend, keyringDir, cmd.InOrStdin(), cdc)
			if err != nil {
				return err
			}

			clientCtx := client.Context{}.
				WithKeyring(kr).
				WithCodec(cdc)

			if err := client.SetCmdClientContextHandler(clientCtx, cmd); err != nil {
				return err
			}

			return nil
		},
	}

	keyringCmd.AddCommand(
		keys.Commands()...,
	)
	keyringCmd.PersistentFlags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|kwallet|pass|test|memory)")

	return keyringCmd
}

func ChainCmd(name string) *cobra.Command {
	keys := keys.Commands()

	keyringCmd := &cobra.Command{
		Use:   "keys",
		Short: fmt.Sprintf("Keyring management for %s chain", name),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cmd.SetContext(context.Background())
			backend, err := cmd.Flags().GetString(flags.FlagKeyringBackend)
			if err != nil {
				return err
			}

			registry := codectypes.NewInterfaceRegistry()
			cdc := codec.NewProtoCodec(registry)

			keyringDir := path.Join(config.DefaultConfigDirName, "keyring", name)
			kr, err := keyring.New(sdk.KeyringServiceName(), backend, keyringDir, cmd.InOrStdin(), cdc)
			if err != nil {
				return err
			}

			clientCtx := client.Context{}.
				WithKeyring(kr).
				WithCodec(cdc)

			if err := client.SetCmdClientContextHandler(clientCtx, cmd); err != nil {
				return err
			}

			return nil
		},
	}

	keyringCmd.AddCommand(
		keys.Commands()...,
	)

	keyringCmd.PersistentFlags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|kwallet|pass|test|memory)")

	return keyringCmd
}
