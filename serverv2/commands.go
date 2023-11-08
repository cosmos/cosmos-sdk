package serverv2

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func Commands(services ...Service) ([]*cobra.Command, error) {
	if len(services) == 0 {
		// TODO figure if we should define default services
		// and if so it should be done here to avoid uncessary dependencies
		return nil, errors.New("no services provided")
	}
	server := NewServer(services...)
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

			defer func() {
				logger.Info("shutting down")
				server.Stop()
			}()

			logger.Info("starting servers...")
			return server.Start()
		},
	}

	return append(server.CLICommands(), startCmd), nil
}
