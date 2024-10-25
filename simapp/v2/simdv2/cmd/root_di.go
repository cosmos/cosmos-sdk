package cmd

import (
	"github.com/spf13/cobra"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/runtime/v2"
	serverv2 "cosmossdk.io/server/v2"
	"cosmossdk.io/simapp/v2"

	"github.com/cosmos/cosmos-sdk/client"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
)

func NewRootCmd[T transaction.Tx](
	args ...string,
) (*cobra.Command, error) {
	rootCommand := &cobra.Command{
		Use:           "simdv2",
		SilenceErrors: true,
	}
	configWriter, err := initRootCmd(rootCommand, log.NewNopLogger(), commandDependencies[T]{})
	if err != nil {
		return nil, err
	}
	factory, err := serverv2.NewCommandFactory(
		serverv2.WithConfigWriter(configWriter),
		serverv2.WithStdDefaultHomeDir(".simappv2"),
		serverv2.WithLoggerFactory(serverv2.NewLogger),
	)
	if err != nil {
		return nil, err
	}

	// returns the target subcommand and a fully realized config map
	subCommand, configMap, logger, err := factory.ParseCommand(rootCommand, args)
	if err != nil {
		return nil, err
	}

	var (
		autoCliOpts     autocli.AppOptions
		moduleManager   *runtime.MM[T]
		clientCtx       client.Context
		simApp          *simapp.SimApp[T]
		depinjectConfig = depinject.Configs(
			depinject.Configs(
				depinject.Supply(logger, runtime.GlobalConfig(configMap)),
				depinject.Provide(ProvideClientContext),
			),
		)
	)
	if isAppRequired(subCommand) {
		// server construction
		simApp, err = simapp.NewSimApp[T](depinjectConfig, &autoCliOpts, &moduleManager, &clientCtx)
		if err != nil {
			return nil, err
		}
	} else {
		// client construction
		if err = depinject.Inject(
			depinject.Configs(
				simapp.AppConfig(),
				depinjectConfig,
			),
			&autoCliOpts, &moduleManager, &clientCtx,
		); err != nil {
			return nil, err
		}
	}

	commandDeps := commandDependencies[T]{
		globalAppConfig: configMap,
		txConfig:        clientCtx.TxConfig,
		moduleManager:   moduleManager,
		simApp:          simApp,
	}
	rootCommand = &cobra.Command{
		Use:               "simdv2",
		Short:             "simulation app",
		SilenceErrors:     true,
		PersistentPreRunE: rootCommandPersistentPreRun(clientCtx),
	}
	factory.EnhanceCommandContext(rootCommand)
	_, err = initRootCmd(rootCommand, logger, commandDeps)
	if err != nil {
		return nil, err
	}
	nodeCmds := nodeservice.NewNodeCommands()
	autoCliOpts.ModuleOptions = make(map[string]*autocliv1.ModuleOptions)
	autoCliOpts.ModuleOptions[nodeCmds.Name()] = nodeCmds.AutoCLIOptions()
	if err := autoCliOpts.EnhanceRootCommand(rootCommand); err != nil {
		return nil, err
	}

	return rootCommand, nil
}
