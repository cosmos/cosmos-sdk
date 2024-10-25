package cmd

import (
	"github.com/spf13/cobra"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/core/server"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/runtime/v2"
	serverv2 "cosmossdk.io/server/v2"
	"cosmossdk.io/simapp/v2"

	"github.com/cosmos/cosmos-sdk/client"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
)

func NewRootCmd(
	commandFixture serverv2.CommandFixture,
	args ...string,
) (*cobra.Command, error) {
	builder, err := serverv2.NewRootCmdBuilder(commandFixture, "simdv2", ".simappv2")
	if err != nil {
		return nil, err
	}
	return builder.Build(args)
}

type DefaultCommandFixture[T transaction.Tx] struct{}

func (DefaultCommandFixture[T]) Bootstrap(cmd *cobra.Command) (serverv2.WritesConfig, error) {
	return initRootCmd(cmd, log.NewNopLogger(), commandDependencies[T]{})
}

func (DefaultCommandFixture[T]) RootCommand(
	rootCommand *cobra.Command,
	subCommand *cobra.Command,
	logger log.Logger,
	configMap server.ConfigMap,
) (*cobra.Command, error) {
	var (
		autoCliOpts   autocli.AppOptions
		moduleManager *runtime.MM[T]
		clientCtx     client.Context
		simApp        *simapp.SimApp[T]
		err           error
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

	rootCommand.Short = "simulation app"
	rootCommand.PersistentPreRunE = rootCommandPersistentPreRun(clientCtx)

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
