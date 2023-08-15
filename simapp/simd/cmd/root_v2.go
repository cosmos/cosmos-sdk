//go:build !app_v1

package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"cosmossdk.io/client/v2/autocli"
	clientv2keyring "cosmossdk.io/client/v2/autocli/keyring"
	"cosmossdk.io/core/address"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/simapp"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	clientrpc "github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	txmodule "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// NewRootCmd creates a new root command for simd. It is called once in the main function.
func NewRootCmd() *cobra.Command {
	var (
		interfaceRegistry  codectypes.InterfaceRegistry
		appCodec           codec.Codec
		txConfig           client.TxConfig
		autoCliOpts        autocli.AppOptions
		moduleBasicManager module.BasicManager
		initClientCtx      client.Context
	)

	if err := depinject.Inject(
		depinject.Configs(simapp.AppConfig,
			depinject.Supply(
				log.NewNopLogger(),
				simtestutil.NewAppOptionsWithFlagHome(tempDir()),
			),
			depinject.Provide(
				ProvideClientContext,
				ProvideKeyring,
			),
		),
		&interfaceRegistry,
		&appCodec,
		&txConfig,
		&autoCliOpts,
		&moduleBasicManager,
		&initClientCtx,
	); err != nil {
		panic(err)
	}

	rootCmd := &cobra.Command{
		Use:           "simd",
		Short:         "simulation app",
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// set the default command outputs
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())

			initClientCtx = initClientCtx.WithCmdContext(cmd.Context())
			initClientCtx, err := client.ReadPersistentCommandFlags(initClientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			initClientCtx, err = config.ReadFromClientConfig(initClientCtx)
			if err != nil {
				return err
			}

			// This needs to go after ReadFromClientConfig, as that function
			// sets the RPC client needed for SIGN_MODE_TEXTUAL.
			enabledSignModes := append(tx.DefaultSignModes, signing.SignMode_SIGN_MODE_TEXTUAL)
			txConfigOpts := tx.ConfigOptions{
				EnabledSignModes:           enabledSignModes,
				TextualCoinMetadataQueryFn: txmodule.NewGRPCCoinMetadataQueryFn(initClientCtx),
			}
			txConfigWithTextual, err := tx.NewTxConfigWithOptions(
				codec.NewProtoCodec(interfaceRegistry),
				txConfigOpts,
			)
			if err != nil {
				return err
			}
			initClientCtx = initClientCtx.WithTxConfig(txConfigWithTextual)
			if err := client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}

			customAppTemplate, customAppConfig := initAppConfig()
			customCMTConfig := initCometBFTConfig()

			return server.InterceptConfigsPreRunHandler(cmd, customAppTemplate, customAppConfig, customCMTConfig)
		},
	}

	initRootCmd(rootCmd, txConfig, interfaceRegistry, appCodec, moduleBasicManager)

	cometCMDs := clientrpc.NewCometBFTCommands()
	autoCliOpts.Modules[cometCMDs.Name()] = cometCMDs

	if err := autoCliOpts.EnhanceRootCommand(rootCmd); err != nil {
		panic(err)
	}

	return rootCmd
}

func ProvideClientContext(appCodec codec.Codec, interfaceRegistry codectypes.InterfaceRegistry, legacyAmino *codec.LegacyAmino) client.Context {
	initClientCtx := client.Context{}.
		WithCodec(appCodec).
		WithInterfaceRegistry(interfaceRegistry).
		WithLegacyAmino(legacyAmino).
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithHomeDir(simapp.DefaultNodeHome).
		WithViper("") // In simapp, we don't use any prefix for env variables.

	// Read the config again to overwrite the default values with the values from the config file
	initClientCtx, _ = config.ReadFromClientConfig(initClientCtx)

	return initClientCtx
}

func ProvideKeyring(clientCtx client.Context, addressCodec address.Codec) (clientv2keyring.Keyring, error) {
	kb, err := client.NewKeyringFromBackend(clientCtx, clientCtx.Keyring.Backend())
	if err != nil {
		return nil, err
	}

	return kb, nil
}
