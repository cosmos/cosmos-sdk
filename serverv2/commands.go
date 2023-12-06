package serverv2

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"cosmossdk.io/log"
	"github.com/spf13/cobra"
)

func Commands(homePath string, modules ...Module) (CLIConfig, error) {
	if len(modules) == 0 {
		// TODO figure if we should define default modules
		// and if so it should be done here to avoid uncessary dependencies
		return CLIConfig{}, errors.New("no modules provided")
	}

	server := NewServer(log.NewLogger(os.Stdout), modules...)
	v, err := server.Config(filepath.Join(homePath, "config"))
	if err != nil {
		return CLIConfig{}, fmt.Errorf("failed to build server config: %w", err)
	}

	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Run the application",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := v.BindPFlags(cmd.Flags()); err != nil {
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
	cmds.Command = append(cmds.Command, startCmd)

	return cmds, nil
}
