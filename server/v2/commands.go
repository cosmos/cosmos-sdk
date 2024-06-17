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

	"cosmossdk.io/log"
)

func Commands(logger log.Logger, homePath string, modules ...ServerModule) (CLIConfig, error) {
	if len(modules) == 0 {
		// TODO figure if we should define default modules
		// and if so it should be done here to avoid unnecessary dependencies
		return CLIConfig{}, errors.New("no modules provided")
	}

	v, err := ReadConfig(filepath.Join(homePath, "config"))
	if err != nil {
		return CLIConfig{}, fmt.Errorf("failed to read config: %w", err)
	}

	server := NewServer(logger, modules...)
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Run the application",
		RunE: func(cmd *cobra.Command, args []string) error {
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

func AddCommands(rootCmd *cobra.Command, logger log.Logger, homePath string, modules ...ServerModule) error {
	cmds, err := Commands(logger, homePath, modules...)
	if err != nil {
		return err
	}

	rootCmd.AddCommand(cmds.Commands...)
	return nil
}
