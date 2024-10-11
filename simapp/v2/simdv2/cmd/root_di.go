package cmd

import (
	serverv2 "cosmossdk.io/server/v2"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/client/v2/autocli"
	clientv2helpers "cosmossdk.io/client/v2/helpers"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/registry"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/runtime/v2"
	"cosmossdk.io/simapp/v2"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtxconfig "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

/*
logger cannot be injected until command line arguments are parsed but injection is needed before
 command line args are parsed due to closing over DI outputs.

Q: Where is logger needed in injection?
A: ProvideEnvironment and ProvideModuleManager need it.  They are receiving a noop logger in the initial injection
and a parsed and configured one in the second.

!! Return a bootstrap command first with Persistent flags, then parse them. This configures the logger and the home directory.
this can also be folded in and nuked:
https://github.com/cosmos/cosmos-sdk/blob/6708818470826923b96ff7fb6ef55729d8c4269e/client/v2/helpers/home.go#L17

*/

// NewRootCmd creates a new root command for simd. It is called once in the main function.
func NewRootCmd[T transaction.Tx](args []string) (*cobra.Command, error) {
	var (
		autoCliOpts   autocli.AppOptions
		moduleManager *runtime.MM[T]
		clientCtx     client.Context
	)
	defaultHomeDir, err := clientv2helpers.DefaultHomeDir(".simappv2")
	if err != nil {
		return nil, err
	}
	bootstrapCmd := &cobra.Command{}
	bootstrapCmd.FParseErrWhitelist.UnknownFlags = true
	bootstrapFlags := bootstrapCmd.PersistentFlags()
	serverv2.SetPersistentFlags(bootstrapFlags, defaultHomeDir)
	err = bootstrapCmd.ParseFlags(args)
	if err != nil {
		return nil, err
	}

	if err = depinject.Inject(
		depinject.Configs(
			simapp.AppConfig(),
			depinject.Provide(ProvideClientContext),
			depinject.Supply(log.NewNopLogger()),
		),
		&autoCliOpts,
		&moduleManager,
		&clientCtx,
	); err != nil {
		return nil, err
	}

	rootCmd := &cobra.Command{
		Use:           "simdv2",
		Short:         "simulation app",
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("simd root command persistent pre-run %+v\n", args)
			clientCtx = clientCtx.WithCmdContext(cmd.Context())
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

			return nil
		},
	}

	initRootCmd(rootCmd, clientCtx.TxConfig, moduleManager)

	nodeCmds := nodeservice.NewNodeCommands()
	autoCliOpts.ModuleOptions = make(map[string]*autocliv1.ModuleOptions)
	autoCliOpts.ModuleOptions[nodeCmds.Name()] = nodeCmds.AutoCLIOptions()

	if err := autoCliOpts.EnhanceRootCommand(rootCmd); err != nil {
		panic(err)
	}

	return rootCmd
}

func ProvideClientContext(
	appCodec codec.Codec,
	interfaceRegistry codectypes.InterfaceRegistry,
	txConfigOpts tx.ConfigOptions,
	legacyAmino registry.AminoRegistrar,
	addressCodec address.Codec,
	validatorAddressCodec address.ValidatorAddressCodec,
	consensusAddressCodec address.ConsensusAddressCodec,
) client.Context {
	var err error

	amino, ok := legacyAmino.(*codec.LegacyAmino)
	if !ok {
		panic("registry.AminoRegistrar must be an *codec.LegacyAmino instance for legacy ClientContext")
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
		WithViper("") // uses by default the binary name as prefix

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
