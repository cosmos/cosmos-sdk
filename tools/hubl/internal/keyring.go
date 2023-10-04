package internal

import (
	"bufio"
	"context"
	"fmt"
	"path"

	"github.com/spf13/cobra"

	_ "cosmossdk.io/api/cosmos/crypto/ed25519"
	_ "cosmossdk.io/api/cosmos/crypto/secp256k1"
	_ "cosmossdk.io/api/cosmos/crypto/secp256r1"
	"cosmossdk.io/tools/hubl/internal/config"
	"cosmossdk.io/tools/hubl/internal/flags"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdkkeyring "github.com/cosmos/cosmos-sdk/crypto/keyring"
)

func getKeyring(chainName string) (sdkkeyring.Keyring, error) {
	if chainName == "" {
		chainName = config.GlobalKeyringDirName
	}

	registry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	configDir, err := config.GetConfigDir()
	if err != nil {
		return nil, err
	}

	cfg, err := config.Load(configDir)
	if err != nil {
		return nil, err
	}

	backend, err := cfg.GetKeyringBackend(chainName)
	if err != nil {
		return nil, err
	}

	keyringDir := path.Join(configDir, "keyring", chainName)
	return sdkkeyring.New(chainName, backend, keyringDir, nil, cdc)
}

func KeyringCmd(chainName string) *cobra.Command {
	shortDesc := fmt.Sprintf("Keyring management for %s", chainName)
	if chainName == "" {
		chainName = config.GlobalKeyringDirName
		shortDesc = "Global keyring management for Hubl"
	}

	keyringCmd := &cobra.Command{
		Use:   "keys",
		Short: shortDesc,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			registry := codectypes.NewInterfaceRegistry()
			cryptocodec.RegisterInterfaces(registry)
			cdc := codec.NewProtoCodec(registry)

			configDir, err := config.GetConfigDir()
			if err != nil {
				return err
			}

			cfg, err := config.Load(configDir)
			if err != nil {
				return err
			}

			backend, err := cfg.GetKeyringBackend(chainName)
			if err != nil {
				return err
			}

			if changed := cmd.Flags().Changed(flags.FlagKeyringBackend); changed {
				b, err := cmd.Flags().GetString(flags.FlagKeyringBackend)
				if err != nil {
					return err
				}

				backend = b
			}

			keyringDir := path.Join(configDir, "keyring", chainName)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			kr, err := sdkkeyring.New(chainName, backend, keyringDir, inBuf, cdc)
			if err != nil {
				return err
			}

			addressCodec, validatorAddressCodec, consensusAddressCodec, err := getAddressCodecFromConfig(cfg, chainName)
			if err != nil {
				return err
			}

			clientCtx := client.Context{}.
				WithKeyring(kr).
				WithCodec(cdc).
				WithKeyringDir(keyringDir).
				WithInput(inBuf).
				WithAddressCodec(addressCodec).
				WithValidatorAddressCodec(validatorAddressCodec).
				WithConsensusAddressCodec(consensusAddressCodec)

			cmd.SetContext(context.WithValue(context.Background(), client.ClientContextKey, &clientCtx))
			if err := client.SetCmdClientContext(cmd, clientCtx); err != nil {
				return err
			}

			return nil
		},
	}

	keyringCmd.AddCommand(
		keys.AddKeyCommand(),
		keys.DeleteKeyCommand(),
		keys.ExportKeyCommand(),
		keys.ImportKeyCommand(),
		keys.ImportKeyHexCommand(),
		keys.ListKeysCmd(),
		keys.ParseKeyStringCommand(),
		keys.RenameKeyCommand(),
		keys.ShowKeysCmd(),
	)
	keyringCmd.PersistentFlags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|kwallet|pass|test|memory)")
	keyringCmd.PersistentFlags().String(flags.FlagOutput, flags.OutputFormatText, "Output format (text|json)")

	return keyringCmd
}
