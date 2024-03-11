//go:build app_v1

package cmd

import (
	"os"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/spf13/cobra"

	"cosmossdk.io/log"
	"cosmossdk.io/simapp"
	"cosmossdk.io/simapp/params"
	"cosmossdk.io/x/auth/tx"
	authtxconfig "cosmossdk.io/x/auth/tx/config"
	"cosmossdk.io/x/auth/types"
	txsigning "cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

// NewRootCmd creates a new root command for simd. It is called once in the
// main function.
func NewRootCmd() *cobra.Command {
	// we "pre"-instantiate the application for getting the injected/configured encoding configuration
	// note, this is not necessary when using app wiring, as depinject can be directly used (see root_v2.go)
	tempApp := simapp.NewSimApp(log.NewNopLogger(), dbm.NewMemDB(), nil, true, simtestutil.NewAppOptionsWithFlagHome(tempDir()))
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

			customClientTemplate, customClientConfig := initClientConfig()
			initClientCtx, err = config.CreateClientConfig(initClientCtx, customClientTemplate, customClientConfig)
			if err != nil {
				return err
			}

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
				return err
			}

			customAppTemplate, customAppConfig := initAppConfig()
			customCMTConfig := initCometBFTConfig()

			return server.InterceptConfigsPreRunHandler(cmd, customAppTemplate, customAppConfig, customCMTConfig)
		},
	}

	initRootCmd(rootCmd, encodingConfig.TxConfig, encodingConfig.InterfaceRegistry, encodingConfig.Codec, tempApp.ModuleManager)

	// autocli opts
	customClientTemplate, customClientConfig := initClientConfig()
	var err error
	initClientCtx, err = config.ReadDefaultValuesFromDefaultClientConfig(initClientCtx, customClientTemplate, customClientConfig)
	if err != nil {
		panic(err)
	}

	autoCliOpts := tempApp.AutoCliOpts()
	autoCliOpts.Keyring, _ = keyring.NewAutoCLIKeyring(initClientCtx.Keyring)
	autoCliOpts.ClientCtx = initClientCtx

	if err := autoCliOpts.EnhanceRootCommand(rootCmd); err != nil {
		panic(err)
	}

	return rootCmd
}
