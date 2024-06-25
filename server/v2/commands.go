package serverv2

import (
	"context"
	"errors"
	"fmt"
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

func Commands(rootCmd *cobra.Command, newApp AppCreator[transaction.Tx], logger log.Logger, components ...ServerComponent[transaction.Tx]) (CLIConfig, error) {
	if len(components) == 0 {
		return CLIConfig{}, errors.New("no components provided")
	}

	server := NewServer(logger, components...)
	flags := server.StartFlags()

	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Run the application",
		RunE: func(cmd *cobra.Command, args []string) error {
			v := GetViperFromCmd(cmd)
			l := GetLoggerFromCmd(cmd)

			for _, startFlags := range flags {
				if err := v.BindPFlags(startFlags); err != nil {
					return err
				}
			}

			if err := v.BindPFlags(cmd.Flags()); err != nil {
				return err
			}

			app := newApp(l, v)

			if err := server.Init(app, v, l); err != nil {
				return err
			}

			srvConfig := Config{StartBlock: true}
			ctx := cmd.Context()
			ctx = context.WithValue(ctx, ServerContextKey, srvConfig)
			ctx, cancelFn := context.WithCancel(ctx)
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
				return fmt.Errorf("failed to start servers: %w", err)
			}

			return nil
		},
	}

	cmds := server.CLICommands()
	cmds.Commands = append(cmds.Commands, startCmd)

	return cmds, nil
}

func AddCommands(rootCmd *cobra.Command, newApp AppCreator[transaction.Tx], logger log.Logger, components ...ServerComponent[transaction.Tx]) error {
	cmds, err := Commands(rootCmd, newApp, logger, components...)
	if err != nil {
		return err
	}

	server := NewServer(logger, components...)
	originalPersistentPreRunE := rootCmd.PersistentPreRunE
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		home, err := cmd.Flags().GetString(FlagHome)
		if err != nil {
			return err
		}

		if err = configHandle(server, home, cmd); err != nil {
			return err
		}

		if rootCmd.PersistentPreRun != nil {
			rootCmd.PersistentPreRun(cmd, args)
			return nil
		}

		return originalPersistentPreRunE(cmd, args)
	}

	rootCmd.AddCommand(cmds.Commands...)
	return nil
}

// configHandle writes the default config to the home directory if it does not exist and sets the server context
func configHandle(s *Server, home string, cmd *cobra.Command) error {
	configDir := filepath.Join(home, "config")

	// we need to check app.toml as the config folder can already exist for the client.toml
	if _, err := os.Stat(filepath.Join(configDir, "app.toml")); os.IsNotExist(err) {
		if err = s.WriteConfig(configDir); err != nil {
			return err
		}
	}

	viper, err := ReadConfig(configDir)
	if err != nil {
		return err
	}
	viper.Set(FlagHome, home)
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	log, err := NewLogger(viper, cmd.OutOrStdout())
	if err != nil {
		return err
	}

	return SetCmdServerContext(cmd, viper, log)
}
