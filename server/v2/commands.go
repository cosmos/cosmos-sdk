package serverv2

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	"github.com/spf13/viper"
)

const flagHome = "home"

type App[T transaction.Tx] struct {
	Application Application[T]
	Store       any
}

type AppCreator[T transaction.Tx] func(*viper.Viper, log.Logger) App[T]

func Commands(rootCmd *cobra.Command, newApp AppCreator[transaction.Tx], logger log.Logger, components ...ServerComponent[transaction.Tx]) (CLIConfig, error) {
	if len(components) == 0 {
		// TODO figure if we should define default components
		// and if so it should be done here to avoid uncessary dependencies
		return CLIConfig{}, errors.New("no components provided")
	}

	server := NewServer(logger, components...)
	// Write default config for each server module
	flags := server.StartFlags()

	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Run the application",
		RunE: func(cmd *cobra.Command, args []string) error {
			v := GetViperFromCmd(cmd)
			l := GetLoggerFromCmd(cmd)

			app := newApp(v, l)
			server.Init(app, v, l)

			for _, startFlags := range flags {
				v.BindPFlags(startFlags)
			}

			if err := v.BindPFlags(cmd.Flags()); err != nil { // the server components are already instantiated here, so binding the flags is useless.
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
		home, err := cmd.Flags().GetString(flagHome)
		if err != nil {
			return err
		}

		err = configHandle(server, home, cmd)
		if err != nil {
			return err
		}

		if originalPersistentPreRunE != nil {
			originalPersistentPreRunE(cmd, args)
			return nil
		}

		return nil
	}

	rootCmd.AddCommand(cmds.Commands...)
	return nil
}

func configHandle(s *Server, home string, cmd *cobra.Command) error {
	if _, err := os.Stat(filepath.Join(home, "config")); os.IsNotExist(err) {
		err = s.WriteConfig(filepath.Join(home, "config"))
		if err != nil {
			return err
		}
	}

	viper, err := ReadConfig(filepath.Join(home, "config"))
	if err != nil {
		return err
	}
	viper.Set(flagHome, home)

	log, err := NewLogger(viper, cmd.OutOrStdout())
	if err != nil {
		return err
	}

	return SetCmdServerContext(cmd, viper, log)
}
