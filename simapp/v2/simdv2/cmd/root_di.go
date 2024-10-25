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
	cmd := &cobra.Command{Use: "simdv2", SilenceErrors: true}
	configWriter, err := initRootCmd(cmd, log.NewNopLogger(), commandDependencies[T]{})
	if err != nil {
		return nil, err
	}
	stdHomeDirOption := serverv2.WithStdDefaultHomeDir(".simappv2")
	factory, err := serverv2.NewCommandFactory(serverv2.WithConfigWriter(configWriter), stdHomeDirOption)
	if err != nil {
		return nil, err
	}

	// returns the target subcommand and a fully realized config map
	subCommand, configMap, err := factory.ParseCommand(cmd, args)
	if err != nil {
		return nil, err
	}
	// create default logger
	logger, err := serverv2.NewLogger(configMap, cmd.OutOrStdout())
	if err != nil {
		return nil, err
	}

	var (
		autoCliOpts   autocli.AppOptions
		moduleManager *runtime.MM[T]
		clientCtx     client.Context
		simApp        *simapp.SimApp[T]
	)
	if isAppRequired(subCommand) {
		// server construction
		simApp, err = simapp.NewSimApp[T](
			depinject.Configs(
				depinject.Supply(logger, runtime.GlobalConfig(configMap)),
				depinject.Provide(ProvideClientContext),
			),
			&autoCliOpts, &moduleManager, &clientCtx)
		if err != nil {
			return nil, err
		}
	} else {
		// client construction
		if err = depinject.Inject(
			depinject.Configs(
				simapp.AppConfig(),
				depinject.Provide(ProvideClientContext),
				depinject.Supply(
					logger,
					runtime.GlobalConfig(configMap),
				),
			),
			&autoCliOpts, &moduleManager, &clientCtx,
		); err != nil {
			return nil, err
		}
	}

	rootCommand := &cobra.Command{
		Use:               "simdv2",
		Short:             "simulation app",
		PersistentPreRunE: rootCommandPersistentPreRun(clientCtx),
	}
	factory, err = serverv2.NewCommandFactory(stdHomeDirOption, serverv2.WithLogger(logger))
	if err != nil {
		return nil, err
	}
	factory.EnhanceCommand(rootCommand)

	commandDeps := commandDependencies[T]{
		globalAppConfig: configMap,
		txConfig:        clientCtx.TxConfig,
		moduleManager:   moduleManager,
		simApp:          simApp,
	}
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
