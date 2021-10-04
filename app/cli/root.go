package cli

import (
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/container"
	"github.com/cosmos/cosmos-sdk/server"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	tmcli "github.com/tendermint/tendermint/libs/cli"
)

type RootCommand struct{ *cobra.Command }

func (RootCommand) IsAutoGroupType() {}

type inputs struct {
	container.In

	RootCommands         []RootCommand
	ClientContextOptions []ClientContextOption
	DefaultHome          DefaultHome
	AppConfig            AppConfig `optional:"true"`
}

func runRoot(in inputs) {
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

			initClientCtx, err = client.ReadPersistentCommandFlags(initClientCtx, cmd.Flags())
			if err != nil {
				return err
			}

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
		tmcli.NewCompletionCmd(rootCmd, true),
		debug.Cmd(),
		config.Cmd(),
	)

	// add keybase, auxiliary RPC, query, and tx child commands
	rootCmd.AddCommand(
		rpc.StatusCommand(),
		keys.Commands(defaultHome),
	)

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
