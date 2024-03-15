//go:build !app_v1

package cmd

import (
	"os"

	"github.com/spf13/cobra"

	authv1 "cosmossdk.io/api/cosmos/auth/module/v1"
	stakingv1 "cosmossdk.io/api/cosmos/staking/module/v1"
	"cosmossdk.io/client/v2/autocli"
	clientv2keyring "cosmossdk.io/client/v2/autocli/keyring"
	"cosmossdk.io/core/address"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/simapp"
	"cosmossdk.io/x/auth/tx"
	authtxconfig "cosmossdk.io/x/auth/tx/config"
	"cosmossdk.io/x/auth/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/types/module"
)

// NewRootCmd creates a new root command for simd. It is called once in the main function.
func NewRootCmd() *cobra.Command {
	var (
		autoCliOpts   autocli.AppOptions
		moduleManager *module.Manager
		clientCtx     client.Context
	)

	if err := depinject.Inject(
		depinject.Configs(simapp.AppConfig(),
			depinject.Supply(
				log.NewNopLogger(),
				simtestutil.NewAppOptionsWithFlagHome(tempDir()),
			),
			depinject.Provide(
				ProvideClientContext,
				ProvideKeyring,
			),
		),
		&autoCliOpts,
		&moduleManager,
		&clientCtx,
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

			clientCtx = clientCtx.WithCmdContext(cmd.Context()).WithViper("")
			clientCtx, err := client.ReadPersistentCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			customClientTemplate, customClientConfig := initClientConfig()
			clientCtx, err = config.CreateClientConfig(clientCtx, customClientTemplate, customClientConfig)
			if err != nil {
				return err
			}

			if err := client.SetCmdClientContextHandler(clientCtx, cmd); err != nil {
				return err
			}

			customAppTemplate, customAppConfig := initAppConfig()
			customCMTConfig := initCometBFTConfig()

			return server.InterceptConfigsPreRunHandler(cmd, customAppTemplate, customAppConfig, customCMTConfig)
		},
	}

	initRootCmd(rootCmd, clientCtx.TxConfig, clientCtx.InterfaceRegistry, clientCtx.Codec, moduleManager)

	if err := autoCliOpts.EnhanceRootCommand(rootCmd); err != nil {
		panic(err)
	}

	return rootCmd
}

func ProvideClientContext(
	appCodec codec.Codec,
	interfaceRegistry codectypes.InterfaceRegistry,
	txConfigOpts tx.ConfigOptions,
	legacyAmino *codec.LegacyAmino,
	addressCodec address.Codec,
	validatorAddressCodec address.ValidatorAddressCodec,
	consensusAddressCodec address.ConsensusAddressCodec,
	authConfig *authv1.Module,
	stakingConfig *stakingv1.Module,
) client.Context {
	var err error

	clientCtx := client.Context{}.
		WithCodec(appCodec).
		WithInterfaceRegistry(interfaceRegistry).
		WithLegacyAmino(legacyAmino).
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithAddressCodec(addressCodec).
		WithValidatorAddressCodec(validatorAddressCodec).
		WithConsensusAddressCodec(consensusAddressCodec).
		WithHomeDir(simapp.DefaultNodeHome).
		WithViper(""). // uses by default the binary name as prefix
		WithAddressPrefix(authConfig.Bech32Prefix).
		WithValidatorPrefix(stakingConfig.Bech32PrefixValidator)

	// Read the config to overwrite the default values with the values from the config file
	customClientTemplate, customClientConfig := initClientConfig()
	clientCtx, err = config.ReadDefaultValuesFromDefaultClientConfig(clientCtx, customClientTemplate, customClientConfig)
	if err != nil {
		panic(err)
	}

	// textual is enabled by default, we need to re-create the tx config grpc instead of bank keeper.
	txConfigOpts.TextualCoinMetadataQueryFn = authtxconfig.NewGRPCCoinMetadataQueryFn(clientCtx)
	txConfig, err := tx.NewTxConfigWithOptions(clientCtx.Codec, txConfigOpts)
	if err != nil {
		panic(err)
	}
	clientCtx = clientCtx.WithTxConfig(txConfig)

	return clientCtx
}

func ProvideKeyring(clientCtx client.Context, addressCodec address.Codec) (clientv2keyring.Keyring, error) {
	kb, err := client.NewKeyringFromBackend(clientCtx, clientCtx.Keyring.Backend())
	if err != nil {
		return nil, err
	}

	return keyring.NewAutoCLIKeyring(kb)
}
