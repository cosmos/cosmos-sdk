package serverv2

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func Commands(modules ...Module) ([]*cobra.Command, error) {
	if len(modules) == 0 {
		// TODO figure if we should define default modules
		// and if so it should be done here to avoid uncessary dependencies
		return nil, errors.New("no modules provided")
	}
	server := NewServer(modules...)
	v, err := server.Configs()
	if err != nil {
		return nil, fmt.Errorf("failed to build server config: %w", err)
	}

	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Run the full node",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := v.BindPFlags(cmd.Flags()); err != nil {
				return err
			}

			logger, err := CreateSDKLogger(v, cmd.OutOrStdout())
			if err != nil {
				return err
			}

			ctx, cleanupFn := context.WithCancel(cmd.Context())
			g, ctx := errgroup.WithContext(ctx)

			ListenForQuitSignals(g, cleanupFn, logger)

			g.Go(func() error {
				logger.Info("starting servers...")
				if err := server.Start(ctx); err != nil {
					return fmt.Errorf("failed to start servers: %w", err)
				}

				// Wait for the calling process to be canceled or close the provided context.
				<-ctx.Done()

				logger.Info("shutting down servers...")
				return server.Stop(ctx)
			})

			return g.Wait()
		},
	}

	return append(server.CLICommands(), startCmd), nil
}
