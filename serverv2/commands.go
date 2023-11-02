package serverv2

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
)

func ServerCmd(services ...Service) (*cobra.Command, error) {
	if len(services) == 0 {
		// TODO figure if we should define default services
		// and if so it should be done here to avoid uncessary dependencies
		return nil, errors.New("no services provided")
	}

	cfg := NewConfig()
	logger, err := CreateSDKLogger(cfg.Viper, os.Stdout)
	if err != nil {
		return nil, err
	}

	server := NewServer(logger, services...)

	baseCmd := &cobra.Command{
		Use:   "server",
		Short: "Experimental server v2",
	}

	baseCmd.AddCommand(startCmd(server))
	baseCmd.AddCommand(server.CLICommands()...)

	return baseCmd, nil
}

func startCmd(server *Server) *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Run the full node",
		RunE: func(cmd *cobra.Command, args []string) error {
			defer server.Stop()
			return server.Start()
		},
	}
}
