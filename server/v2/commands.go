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
	"github.com/spf13/viper"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
)

type NewServerModulesFunc func(*viper.Viper, log.Logger, transaction.Codec[transaction.Tx]) []ServerModule

func Commands(rootCmd *cobra.Command, appCreator NewServerModulesFunc, codec transaction.Codec[transaction.Tx], logger log.Logger, homePath string, modules ...ServerModule,) (CLIConfig, error) {
	if len(modules) == 0 {
		// TODO figure if we should define default modules
		// and if so it should be done here to avoid uncessary dependencies
		return CLIConfig{}, errors.New("no modules provided")
	}

	server := NewServer(logger, modules...)
	// Write default config for each server module
	flags := server.StartFlags()
	// Check if config folder exist
	if _, err := os.Stat(filepath.Join(homePath, "config")); os.IsNotExist(err) {
		_ = server.WriteConfig(filepath.Join(homePath, "config"))
	}
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Run the application",
		RunE: func(cmd *cobra.Command, args []string) error {
			v := GetViperFromCmd(cmd)
			l := GetLoggerFromCmd(cmd)

			server.modules = appCreator(v, l, codec)

			for _, startFlags := range flags {
				v.BindPFlags(startFlags)
			}
			
			if err := v.BindPFlags(cmd.Flags()); err != nil { // the server modules are already instantiated here, so binding the flags is useless.
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

func AddCommands(rootCmd *cobra.Command, appCreator NewServerModulesFunc, codec transaction.Codec[transaction.Tx], logger log.Logger, homePath string, modules ...ServerModule) error {
	cmds, err := Commands(rootCmd, appCreator, codec, logger, homePath, modules...)
	if err != nil {
		return err
	}

	rootCmd.AddCommand(cmds.Commands...)
	return nil
}
