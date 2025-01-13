package cmd

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/runtime/v2"
	serverv2 "cosmossdk.io/server/v2"
	"cosmossdk.io/simapp/v2"

	"github.com/cosmos/cosmos-sdk/client"
)

func NewRootCmd[T transaction.Tx](
	args ...string,
) (*cobra.Command, error) {
	rootCommand := &cobra.Command{
		Use:           "simdv2",
		SilenceErrors: true,
	}
	configWriter, err := InitRootCmd(rootCommand, log.NewNopLogger(), CommandDependencies[T]{})
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

	var autoCliOpts autocli.AppOptions
	if err := depinject.Inject(
		depinject.Configs(
			simapp.AppConfig(),
			depinject.Supply(runtime.GlobalConfig{}, log.NewNopLogger())),
		&autoCliOpts,
	); err != nil {
		return nil, err
	}

	if err = autoCliOpts.EnhanceRootCommand(rootCommand); err != nil {
		return nil, err
	}
	subCommand, configMap, logger, err := factory.ParseCommand(rootCommand, args)
	if err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			return rootCommand, nil
		}
		return nil, err
	}

	var (
		moduleManager   *runtime.MM[T]
		clientCtx       client.Context
		simApp          *simapp.SimApp[T]
		depinjectConfig = depinject.Configs(
			depinject.Supply(logger, runtime.GlobalConfig(configMap)),
			depinject.Provide(ProvideClientContext),
		)
	)
	if serverv2.IsAppRequired(subCommand) {
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

	commandDeps := CommandDependencies[T]{
		GlobalConfig:  configMap,
		TxConfig:      clientCtx.TxConfig,
		ModuleManager: moduleManager,
		SimApp:        simApp,
		ClientContext: clientCtx,
	}
	rootCommand = &cobra.Command{
		Use:               "simdv2",
		Short:             "simulation app",
		SilenceErrors:     true,
		PersistentPreRunE: RootCommandPersistentPreRun(clientCtx),
	}
	factory.EnhanceRootCommand(rootCommand)
	_, err = InitRootCmd(rootCommand, logger, commandDeps)
	if err != nil {
		return nil, err
	}

	if err := autoCliOpts.EnhanceRootCommand(rootCommand); err != nil {
		return nil, err
	}

	return rootCommand, nil
}
