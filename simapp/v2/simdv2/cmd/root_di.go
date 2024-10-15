package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/core/server"
	"cosmossdk.io/log"
	serverv2 "cosmossdk.io/server/v2"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"cosmossdk.io/client/v2/autocli"
	clientv2helpers "cosmossdk.io/client/v2/helpers"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/registry"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/depinject"
	"cosmossdk.io/runtime/v2"
	"cosmossdk.io/simapp/v2"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
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

DI happens before commands are even created and therefore before full CLI flags binding
In the DI phase only flags mentioned in server.ModuleConfigMap will be available from CLI
In the DI phase viper config is fully available, but not overrides from CLI flags
Server components are invoked on start.
Server components receive fully parse CLI flags since the Start invocation happens in command.RunE

*/

type ModuleConfigMaps map[string]server.ConfigMap
type FlagParser func() error
type GlobalConfig server.ConfigMap

func ProvideModuleConfigMap(
	moduleConfigs []server.ModuleConfigMap,
	flags *pflag.FlagSet,
	parseFlags FlagParser,
) (ModuleConfigMaps, error) {
	var err error
	const bootstrapFlags = "__bootstrap"
	globalConfig := make(ModuleConfigMaps)
	globalConfig[bootstrapFlags] = make(server.ConfigMap)
	for _, moduleConfig := range moduleConfigs {
		cfg := moduleConfig.Config
		name := moduleConfig.Module
		globalConfig[name] = make(server.ConfigMap)
		for flag, defaultValue := range cfg {
			globalConfig[name][flag] = defaultValue
			switch v := defaultValue.(type) {
			case string:
				_, maybeNotFound := flags.GetString(flag)
				if maybeNotFound != nil && strings.Contains(maybeNotFound.Error(),
					"flag accessed but not defined") {
					flags.String(flag, v, "")
				} else {
					// silently skip the flag if it's already defined
					continue
				}
			case []int:
				flags.IntSlice(flag, v, "")
			case int:
				flags.Int(flag, v, "")
			default:
				return nil, fmt.Errorf("unsupported type %T for flag %s", defaultValue, flag)
			}
			if err != nil {
				return nil, err
			}
		}
	}
	if err = parseFlags(); err != nil {
		return nil, err
	}
	for _, cfg := range globalConfig {
		for flag, defaulValue := range cfg {
			var val any
			switch defaulValue.(type) {
			case string:
				val, err = flags.GetString(flag)
			case []int:
				val, err = flags.GetIntSlice(flag)
			case int:
				val, err = flags.GetInt(flag)
			default:
				return nil, fmt.Errorf("unsupported type %T for flag %s", defaulValue, flag)
			}
			if err != nil {
				return nil, err
			}
			cfg[flag] = val
			// also collect all flags into bootstrap config
			globalConfig[bootstrapFlags][flag] = val
		}
	}
	return globalConfig, nil
}

// NewRootCmd creates a new root command for simd. It is called once in the main function.
func NewRootCmd[T transaction.Tx](args []string) (*cobra.Command, error) {
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
	// we need to check app.toml as the config folder can already exist for the client.toml
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
	err = serverv2.SetCmdServerContext(cmd, vipr, logger)
	if err != nil {
		return nil, err
	}
	globalConfig := vipr.AllSettings()

	var app *simapp.SimApp[T]
	if needsApp(cmd) {
		app, err = simapp.NewSimAppWithConfig[T](
			depinject.Configs(
				depinject.Supply(logger, simapp.GlobalConfig(globalConfig)),
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
					simapp.GlobalConfig(globalConfig),
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
	srv, err = initRootCmd[T](rootCmd, logger, globalConfig, clientCtx.TxConfig, moduleManager, app)
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
