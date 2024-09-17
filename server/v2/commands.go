package serverv2

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
)

// Execute executes the root command of an application.
// It handles adding core CLI flags, specifically the logging flags.
func Execute(rootCmd *cobra.Command, envPrefix, defaultHome string) error {
	rootCmd.PersistentFlags().String(FlagLogLevel, "info", "The logging level (trace|debug|info|warn|error|fatal|panic|disabled or '*:<level>,<key>:<level>')")
	rootCmd.PersistentFlags().String(FlagLogFormat, "plain", "The logging format (json|plain)")
	rootCmd.PersistentFlags().Bool(FlagLogNoColor, false, "Disable colored logs")
	rootCmd.PersistentFlags().StringP(FlagHome, "", defaultHome, "directory for config and data")

	// update the global viper with the root command's configuration
	viper.SetEnvPrefix(envPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

	return rootCmd.Execute()
}

// AddCommands add the server commands to the root command
// It configure the config handling and the logger handling
func AddCommands[T transaction.Tx](
	rootCmd *cobra.Command,
	newApp AppCreator[T],
	logger log.Logger,
	globalServerCfg ServerConfig,
	components ...ServerComponent[T],
) error {
	if len(components) == 0 {
		return errors.New("no components provided")
	}

	server := NewServer(logger, globalServerCfg, components...)
	originalPersistentPreRunE := rootCmd.PersistentPreRunE
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// set the default command outputs
		cmd.SetOut(cmd.OutOrStdout())
		cmd.SetErr(cmd.ErrOrStderr())

		if err := configHandle(server, cmd); err != nil {
			return err
		}

		// call the original PersistentPreRun(E) if it exists
		if rootCmd.PersistentPreRun != nil {
			rootCmd.PersistentPreRun(cmd, args)
			return nil
		}

		return originalPersistentPreRunE(cmd, args)
	}

	cmds := server.CLICommands()
	startCmd := createStartCommand(server, newApp)
	startCmd.SetContext(rootCmd.Context())
	cmds.Commands = append(cmds.Commands, startCmd)
	rootCmd.AddCommand(cmds.Commands...)

	if len(cmds.Queries) > 0 {
		if queryCmd := findSubCommand(rootCmd, "query"); queryCmd != nil {
			queryCmd.AddCommand(cmds.Queries...)
		} else {
			queryCmd := topLevelCmd(rootCmd.Context(), "query", "Querying subcommands")
			queryCmd.Aliases = []string{"q"}
			queryCmd.AddCommand(cmds.Queries...)
			rootCmd.AddCommand(queryCmd)
		}
	}

	if len(cmds.Txs) > 0 {
		if txCmd := findSubCommand(rootCmd, "tx"); txCmd != nil {
			txCmd.AddCommand(cmds.Txs...)
		} else {
			txCmd := topLevelCmd(rootCmd.Context(), "tx", "Transactions subcommands")
			txCmd.AddCommand(cmds.Txs...)
			rootCmd.AddCommand(txCmd)
		}
	}

	return nil
}

// createStartCommand creates the start command for the application.
func createStartCommand[T transaction.Tx](
	server *Server[T],
	newApp AppCreator[T],
) *cobra.Command {
	flags := server.StartFlags()

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Run the application",
		RunE: func(cmd *cobra.Command, args []string) error {
			v := GetViperFromCmd(cmd)
			l := GetLoggerFromCmd(cmd)
			if err := v.BindPFlags(cmd.Flags()); err != nil {
				return err
			}

			if err := server.Init(newApp(l, v), v.AllSettings(), l); err != nil {
				return err
			}

			ctx, cancelFn := context.WithCancel(cmd.Context())
			go func() {
				sigCh := make(chan os.Signal, 1)
				signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
				sig := <-sigCh
				cancelFn()
				cmd.Printf("caught %s signal\n", sig.String())

				if err := server.Stop(ctx); err != nil {
					cmd.PrintErrln("failed to stop servers:", err)
				}
			}()

			if err := server.Start(ctx); err != nil {
				return err
			}

			return nil
		},
	}

	// add the start flags to the command
	for _, startFlags := range flags {
		cmd.Flags().AddFlagSet(startFlags)
	}

	return cmd
}

// configHandle writes the default config to the home directory if it does not exist and sets the server context
func configHandle[T transaction.Tx](s *Server[T], cmd *cobra.Command) error {
	home, err := cmd.Flags().GetString(FlagHome)
	if err != nil {
		return err
	}

	configDir := filepath.Join(home, "config")

	// we need to check app.toml as the config folder can already exist for the client.toml
	if _, err := os.Stat(filepath.Join(configDir, "app.toml")); os.IsNotExist(err) {
		if err = s.WriteConfig(configDir); err != nil {
			return err
		}
	}

	v, err := ReadConfig(configDir)
	if err != nil {
		return err
	}

	if err := v.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	log, err := NewLogger(v, cmd.OutOrStdout())
	if err != nil {
		return err
	}

	return SetCmdServerContext(cmd, v, log)
}

// findSubCommand finds a sub-command of the provided command whose Use
// string is or begins with the provided subCmdName.
// It verifies the command's aliases as well.
func findSubCommand(cmd *cobra.Command, subCmdName string) *cobra.Command {
	for _, subCmd := range cmd.Commands() {
		use := subCmd.Use
		if use == subCmdName || strings.HasPrefix(use, subCmdName+" ") {
			return subCmd
		}

		for _, alias := range subCmd.Aliases {
			if alias == subCmdName || strings.HasPrefix(alias, subCmdName+" ") {
				return subCmd
			}
		}
	}
	return nil
}

// topLevelCmd creates a new top-level command with the provided name and
// description. The command will have DisableFlagParsing set to false and
// SuggestionsMinimumDistance set to 2.
func topLevelCmd(ctx context.Context, use, short string) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        use,
		Short:                      short,
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
	}
	cmd.SetContext(ctx)

	return cmd
}
