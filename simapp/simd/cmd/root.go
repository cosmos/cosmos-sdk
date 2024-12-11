<<<<<<< HEAD
//go:build app_v1
=======
//go:build !app_v1
>>>>>>> aa8266e70 (docs: runtime docs (#22816))

package cmd

import (
	"os"

	"github.com/spf13/cobra"

<<<<<<< HEAD
	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/log"
	"cosmossdk.io/simapp"
	"cosmossdk.io/simapp/params"
	txsigning "cosmossdk.io/x/tx/signing"
=======
	authv1 "cosmossdk.io/api/cosmos/auth/module/v1"
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	stakingv1 "cosmossdk.io/api/cosmos/staking/module/v1"
	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/registry"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/simapp"
	basedepinject "cosmossdk.io/x/accounts/defaults/base/depinject"
	lockupdepinject "cosmossdk.io/x/accounts/defaults/lockup/depinject"
	multisigdepinject "cosmossdk.io/x/accounts/defaults/multisig/depinject"
>>>>>>> aa8266e70 (docs: runtime docs (#22816))

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
<<<<<<< HEAD
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/server"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
=======
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/types/module"
>>>>>>> aa8266e70 (docs: runtime docs (#22816))
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtxconfig "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

<<<<<<< HEAD
// NewRootCmd creates a new root command for simd. It is called once in the
// main function.
func NewRootCmd() *cobra.Command {
	// we "pre"-instantiate the application for getting the injected/configured encoding configuration
	// note, this is not necessary when using app wiring, as depinject can be directly used (see root_v2.go)
	tempApp := simapp.NewSimApp(log.NewNopLogger(), coretesting.NewMemDB(), nil, true, simtestutil.NewAppOptionsWithFlagHome(simapp.DefaultNodeHome))
	encodingConfig := params.EncodingConfig{
		InterfaceRegistry: tempApp.InterfaceRegistry(),
		Codec:             tempApp.AppCodec(),
		TxConfig:          tempApp.TxConfig(),
		Amino:             tempApp.LegacyAmino(),
	}

	initClientCtx := client.Context{}.
		WithCodec(encodingConfig.Codec).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithAddressCodec(addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())).
		WithValidatorAddressCodec(addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix())).
		WithConsensusAddressCodec(addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix())).
		WithHomeDir(simapp.DefaultNodeHome).
		WithViper(""). // uses by default the binary name as prefix
		WithAddressPrefix(sdk.GetConfig().GetBech32AccountAddrPrefix()).
		WithValidatorPrefix(sdk.GetConfig().GetBech32ValidatorAddrPrefix())
=======
// NewRootCmd creates a new root command for simd. It is called once in the main function.
func NewRootCmd() *cobra.Command {
	var (
		autoCliOpts   autocli.AppOptions
		moduleManager *module.Manager
		clientCtx     client.Context
	)

	if err := depinject.Inject(
		depinject.Configs(simapp.AppConfig(),
			depinject.Supply(log.NewNopLogger()),
			depinject.Provide(
				ProvideClientContext,
				multisigdepinject.ProvideAccount,
				basedepinject.ProvideAccount,
				lockupdepinject.ProvideAllLockupAccounts,

				// provide base account options
				basedepinject.ProvideSecp256K1PubKey,
			),
		),
		&autoCliOpts,
		&moduleManager,
		&clientCtx,
	); err != nil {
		panic(err)
	}
>>>>>>> aa8266e70 (docs: runtime docs (#22816))

	rootCmd := &cobra.Command{
		Use:           "simd",
		Short:         "simulation app",
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// set the default command outputs
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())

<<<<<<< HEAD
			initClientCtx = initClientCtx.WithCmdContext(cmd.Context())
			initClientCtx, err := client.ReadPersistentCommandFlags(initClientCtx, cmd.Flags())
=======
			clientCtx = clientCtx.WithCmdContext(cmd.Context()).WithViper("")
			clientCtx, err := client.ReadPersistentCommandFlags(clientCtx, cmd.Flags())
>>>>>>> aa8266e70 (docs: runtime docs (#22816))
			if err != nil {
				return err
			}

			customClientTemplate, customClientConfig := initClientConfig()
<<<<<<< HEAD
			initClientCtx, err = config.CreateClientConfig(initClientCtx, customClientTemplate, customClientConfig)
=======
			clientCtx, err = config.CreateClientConfig(clientCtx, customClientTemplate, customClientConfig)
>>>>>>> aa8266e70 (docs: runtime docs (#22816))
			if err != nil {
				return err
			}

<<<<<<< HEAD
			// This needs to go after CreateClientConfig, as that function
			// sets the RPC client needed for SIGN_MODE_TEXTUAL. This sign mode
			// is only available if the client is online.
			if !initClientCtx.Offline {
				enabledSignModes := append(tx.DefaultSignModes, signing.SignMode_SIGN_MODE_TEXTUAL)
				txConfigOpts := tx.ConfigOptions{
					EnabledSignModes:           enabledSignModes,
					TextualCoinMetadataQueryFn: authtxconfig.NewGRPCCoinMetadataQueryFn(initClientCtx),
					SigningOptions: &txsigning.Options{
						AddressCodec:          initClientCtx.InterfaceRegistry.SigningContext().AddressCodec(),
						ValidatorAddressCodec: initClientCtx.InterfaceRegistry.SigningContext().ValidatorAddressCodec(),
					},
				}
				txConfig, err := tx.NewTxConfigWithOptions(
					initClientCtx.Codec,
					txConfigOpts,
				)
				if err != nil {
					return err
				}

				initClientCtx = initClientCtx.WithTxConfig(txConfig)
			}

			if err := client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
=======
			if err := client.SetCmdClientContextHandler(clientCtx, cmd); err != nil {
>>>>>>> aa8266e70 (docs: runtime docs (#22816))
				return err
			}

			customAppTemplate, customAppConfig := initAppConfig()
			customCMTConfig := initCometBFTConfig()

			return server.InterceptConfigsPreRunHandler(cmd, customAppTemplate, customAppConfig, customCMTConfig)
		},
	}

<<<<<<< HEAD
	initRootCmd(rootCmd, tempApp.ModuleManager)

	// autocli opts
	customClientTemplate, customClientConfig := initClientConfig()
	var err error
	initClientCtx, err = config.CreateClientConfig(initClientCtx, customClientTemplate, customClientConfig)
	if err != nil {
		panic(err)
	}

	autoCliOpts := tempApp.AutoCliOpts()
	autoCliOpts.AddressCodec = initClientCtx.AddressCodec
	autoCliOpts.ValidatorAddressCodec = initClientCtx.ValidatorAddressCodec
	autoCliOpts.ConsensusAddressCodec = initClientCtx.ConsensusAddressCodec
	autoCliOpts.Cdc = initClientCtx.Codec

	nodeCmds := nodeservice.NewNodeCommands()
=======
	initRootCmd(rootCmd, moduleManager)

	nodeCmds := nodeservice.NewNodeCommands()
	autoCliOpts.ModuleOptions = make(map[string]*autocliv1.ModuleOptions)
>>>>>>> aa8266e70 (docs: runtime docs (#22816))
	autoCliOpts.ModuleOptions[nodeCmds.Name()] = nodeCmds.AutoCLIOptions()

	if err := autoCliOpts.EnhanceRootCommand(rootCmd); err != nil {
		panic(err)
	}

	return rootCmd
}
<<<<<<< HEAD
=======

func ProvideClientContext(
	appCodec codec.Codec,
	interfaceRegistry codectypes.InterfaceRegistry,
	txConfigOpts tx.ConfigOptions,
	legacyAmino registry.AminoRegistrar,
	addressCodec address.Codec,
	validatorAddressCodec address.ValidatorAddressCodec,
	consensusAddressCodec address.ConsensusAddressCodec,
	authConfig *authv1.Module,
	stakingConfig *stakingv1.Module,
) client.Context {
	var err error

	amino, ok := legacyAmino.(*codec.LegacyAmino)
	if !ok {
		panic("ProvideClientContext requires a *codec.LegacyAmino instance")
	}

	clientCtx := client.Context{}.
		WithCodec(appCodec).
		WithInterfaceRegistry(interfaceRegistry).
		WithLegacyAmino(amino).
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
	clientCtx, err = config.CreateClientConfig(clientCtx, customClientTemplate, customClientConfig)
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
>>>>>>> aa8266e70 (docs: runtime docs (#22816))
