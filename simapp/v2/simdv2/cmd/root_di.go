package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/client/v2/autocli"
	clientv2helpers "cosmossdk.io/client/v2/helpers"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/registry"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/runtime/v2"
	serverv2 "cosmossdk.io/server/v2"
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

// NewRootCmd creates a new root command for simd. It is called once in the main function.
func NewRootCmd[T transaction.Tx](args ...string) (*cobra.Command, error) {
	var (
		autoCliOpts   autocli.AppOptions
		moduleManager *runtime.MM[T]
		clientCtx     client.Context
	)
	// parse config
	defaultHomeDir, err := clientv2helpers.DefaultHomeDir(".simappv2")
	if err != nil {
		return nil, err
	}

	// initial bootstrap root command for config parsing
	rootCmd := &cobra.Command{
		Use:           "simdv2",
		Short:         "simulation app",
		SilenceErrors: true,
	}
	serverv2.SetPersistentFlags(rootCmd.PersistentFlags(), defaultHomeDir)
	// update the global viper with the root command's configuration
	viper.SetEnvPrefix(clientv2helpers.EnvPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

	srv, err := initRootCmd[T](rootCmd, log.NewNopLogger(), nil, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	cmd, _, err := rootCmd.Traverse(args)
	if cmd == nil || err != nil {
		return rootCmd, nil
	}
	if err = cmd.ParseFlags(args); err != nil {
		if err.Error() == "pflag: help requested" {
			return cmd, nil
		}
		return nil, err
	}
	home, err := cmd.Flags().GetString(serverv2.FlagHome)
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(home, "config")
	// create app.toml if it does not already exist
	if _, err := os.Stat(filepath.Join(configDir, "app.toml")); os.IsNotExist(err) {
		if err = srv.WriteConfig(configDir); err != nil {
			return nil, err
		}
	}
	vipr, err := serverv2.ReadConfig(configDir)
	if err != nil {
		return nil, err
	}
	if err := vipr.BindPFlags(cmd.Flags()); err != nil {
		return nil, err
	}
	logger, err := serverv2.NewLogger(vipr, cmd.OutOrStdout())
	if err != nil {
		return nil, err
	}
	globalConfig := vipr.AllSettings()

	var simApp *simapp.SimApp[T]
	if needsApp(cmd) {
		simApp, err = simapp.NewSimApp[T](
			depinject.Configs(
				depinject.Supply(logger, runtime.GlobalConfig(globalConfig)),
				depinject.Provide(ProvideClientContext),
			),
			&autoCliOpts, &moduleManager, &clientCtx)
		if err != nil {
			return nil, err
		}
	} else {
		if err = depinject.Inject(
			depinject.Configs(
				simapp.AppConfig(),
				depinject.Provide(
					ProvideClientContext,
				),
				depinject.Supply(
					logger,
					runtime.GlobalConfig(globalConfig),
				),
			),
			&autoCliOpts,
			&moduleManager,
			&clientCtx,
		); err != nil {
			return nil, err
		}
	}

	// final root command
	rootCmd = &cobra.Command{
		Use:           "simdv2",
		Short:         "simulation app",
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// set the default command outputs
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())

			clientCtx = clientCtx.WithCmdContext(cmd.Context())
			clientCtx, err = client.ReadPersistentCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			customClientTemplate, customClientConfig := initClientConfig()
			clientCtx, err = config.CreateClientConfig(
				clientCtx, customClientTemplate, customClientConfig)
			if err != nil {
				return err
			}

			if err = client.SetCmdClientContextHandler(clientCtx, cmd); err != nil {
				return err
			}

			return nil
		},
	}
	// TODO push down to server subcommands only
	// but also audit the usage of logger and viper in fetching from context, probably no
	// longer needed.
	serverv2.SetPersistentFlags(rootCmd.PersistentFlags(), defaultHomeDir)
	err = serverv2.SetCmdServerContext(rootCmd, vipr, logger)
	if err != nil {
		return nil, err
	}

	_, err = initRootCmd[T](rootCmd, logger, globalConfig, clientCtx.TxConfig, moduleManager, simApp)
	if err != nil {
		return nil, err
	}
	nodeCmds := nodeservice.NewNodeCommands()
	autoCliOpts.ModuleOptions = make(map[string]*autocliv1.ModuleOptions)
	autoCliOpts.ModuleOptions[nodeCmds.Name()] = nodeCmds.AutoCLIOptions()
	if err := autoCliOpts.EnhanceRootCommand(rootCmd); err != nil {
		return nil, err
	}

	return rootCmd, nil
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

func needsApp(cmd *cobra.Command) bool {
	if cmd.Annotations["needs-app"] == "true" {
		return true
	}
	if cmd.Parent() == nil {
		return false
	}
	return needsApp(cmd.Parent())
}
