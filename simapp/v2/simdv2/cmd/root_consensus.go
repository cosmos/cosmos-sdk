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
	"github.com/cosmos/cosmos-sdk/client/config"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
)

// AlternateConsensusFixture is a custom implementation of the CommandFixture interface
// which demonstrates how an app developer can customize the root command initialization.
// In this case a different consensus assummed but not implemented.
type AlternateConsensusFixture[T transaction.Tx] struct{}

func (AlternateConsensusFixture[T]) Bootstrap(cmd *cobra.Command) (serverv2.WritesConfig, error) {
	// create your custom consensus component here
	var consensusComponent serverv2.ServerComponent[T]
	return initRootCmd(cmd, log.NewNopLogger(), commandDependencies[T]{
		consensusComponent: consensusComponent,
	})
}

func (AlternateConsensusFixture[T]) RootCommand(
	rootCommand *cobra.Command,
	subCommand *cobra.Command,
	logger log.Logger,
	configMap server.ConfigMap,
) (*cobra.Command, error) {
	var (
		autoCliOpts        autocli.AppOptions
		moduleManager      *runtime.MM[T]
		clientCtx          client.Context
		simApp             *simapp.SimApp[T]
		err                error
		consensusComponent serverv2.ServerComponent[T]
	)
	if needsApp(subCommand) {
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
		// initialize your custom consensus component here with outputs now available from DI
		consensusComponent = nil
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
	rootCommand.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
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
	}

	commandDeps := commandDependencies[T]{
		globalAppConfig:    configMap,
		txConfig:           clientCtx.TxConfig,
		moduleManager:      moduleManager,
		simApp:             simApp,
		consensusComponent: consensusComponent,
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
