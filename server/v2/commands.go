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
)

func Commands[AppT AppI[T], T transaction.Tx](
	rootCmd *cobra.Command,
	newApp AppCreator[AppT, T],
	logger log.Logger,
	components ...ServerComponent[AppT, T],
) (CLIConfig, error) {
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

func AddCommands[AppT AppI[T], T transaction.Tx](
	rootCmd *cobra.Command,
	newApp AppCreator[AppT, T],
	logger log.Logger,
	components ...ServerComponent[AppT, T],
) error {
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

		err = configHandle(server, home, cmd)
		if err != nil {
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
func configHandle[AppT AppI[T], T transaction.Tx](s *Server[AppT, T], home string, cmd *cobra.Command) error {
	if _, err := os.Stat(filepath.Join(home, "config")); os.IsNotExist(err) {
		if err = s.WriteConfig(filepath.Join(home, "config")); err != nil {
			return err
		}
	}

	viper, err := ReadConfig(filepath.Join(home, "config"))
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
