package cli

import (
	"fmt"
	io "io"
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/spf13/cobra"

	tmcli "github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/container"
	"github.com/cosmos/cosmos-sdk/server"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func Run(options ...container.Option) {
	err := container.Run(runner, options...)
	if err != nil {
		panic(err)
	}
}

type inputs struct {
	RootCommands          []RootCommand
	QueryCommands         []QueryCommand
	TxCommands            []TxCommand
	ClientContextOptions  []ClientContextOption
	DefaultHome           DefaultHome
	AppCreator            types.AppCreator      `optional:"true"`
	AppExporter           types.AppExporter     `optional:"true"`
	StartCommandInitFlags types.ModuleInitFlags `optional:"true"`
	AppConfig             AppConfig             `optional:"true"`
}

func runner(in inputs) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	defaultHome := filepath.Join(userHomeDir, string(in.DefaultHome))

	initClientCtx := client.Context{}.
		WithInput(os.Stdin).
		WithHomeDir(defaultHome)

	for _, opt := range in.ClientContextOptions {
		initClientCtx = opt(initClientCtx)
	}

	rootCmd := &cobra.Command{
		Use:   "TODO",
		Short: "TODO",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// set the default command outputs
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())

			initClientCtx = client.ReadHomeFlag(initClientCtx, cmd)

			initClientCtx, err := config.ReadFromClientConfig(initClientCtx)
			if err != nil {
				return err
			}

			if err := client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}

			appConfig := in.AppConfig.Config
			if appConfig == nil {
				appConfig = serverconfig.DefaultConfig()
			}

			appConfigTemplate := in.AppConfig.Template
			if appConfigTemplate == "" {
				appConfigTemplate = serverconfig.DefaultConfigTemplate
			}

			return server.InterceptConfigsPreRunHandler(cmd, appConfigTemplate, appConfig)
		},
	}

	cfg := sdk.GetConfig()
	cfg.Seal()

	rootCmd.AddCommand(
		//TODO: AddGenesisAccountCmd(simapp.DefaultNodeHome),
		tmcli.NewCompletionCmd(rootCmd, true),
		//TODO: NewTestnetCmd(simapp.ModuleBasics, banktypes.GenesisBalancesIterator{}),
		debug.Cmd(),
		config.Cmd(),
	)

	if in.AppCreator != nil {
		appExporter := in.AppExporter
		if appExporter == nil {
			appExporter = func(logger log.Logger, db dbm.DB, writer io.Writer, i int64, b bool, strings []string, options types.AppOptions) (types.ExportedApp, error) {
				return types.ExportedApp{}, fmt.Errorf("no AppExporter")
			}
		}

		initFlags := in.StartCommandInitFlags
		if initFlags == nil {
			initFlags = func(startCmd *cobra.Command) {}
		}

		server.AddCommands(rootCmd, defaultHome, in.AppCreator, appExporter, initFlags)
	}

	// add keybase, auxiliary RPC, query, and tx child commands
	rootCmd.AddCommand(
		rpc.StatusCommand(),
		keys.Commands(defaultHome),
	)

	// TODO: add rosetta
	// rootCmd.AddCommand(server.RosettaCommand(encodingConfig.InterfaceRegistry, encodingConfig.Codec))

	if len(in.QueryCommands) > 0 {
		rootCmd.AddCommand(makeQueryCommand(in.QueryCommands))
	}

	if len(in.TxCommands) > 0 {
		rootCmd.AddCommand(makeTxCommand(in.TxCommands))
	}

	for _, cmd := range in.RootCommands {
		rootCmd.AddCommand(cmd.Command)
	}

	if err := svrcmd.Execute(rootCmd, defaultHome); err != nil {
		switch e := err.(type) {
		case server.ErrorCode:
			os.Exit(e.Code)

		default:
			os.Exit(1)
		}
	}
}
func makeQueryCommand(commands []QueryCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		rpc.ValidatorCommand(),
		rpc.BlockCommand(),
	)

	for _, c := range commands {
		if c.Command != nil {
			cmd.AddCommand(c.Command)
		}
	}

	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}

func makeTxCommand(commands []TxCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	for _, c := range commands {
		if c.Command != nil {
			cmd.AddCommand(c.Command)
		}
	}

	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}
