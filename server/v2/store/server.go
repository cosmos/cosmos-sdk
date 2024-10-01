package store

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	serverv2 "cosmossdk.io/server/v2"
	storev2 "cosmossdk.io/store/v2"
)

var (
	_ serverv2.ServerComponent[transaction.Tx] = (*Server[transaction.Tx])(nil)
	_ serverv2.HasConfig                       = (*Server[transaction.Tx])(nil)
	_ serverv2.HasCLICommands                  = (*Server[transaction.Tx])(nil)
)

const ServerName = "store"

// Server manages store config and contains prune & snapshot commands
type Server[T transaction.Tx] struct {
	config *Config
	store  storev2.Backend
}

func New[T transaction.Tx](store storev2.Backend) *Server[T] {
	return &Server[T]{store: store}
}

func (s *Server[T]) Init(_ serverv2.AppI[T], cfg map[string]any, _ log.Logger) error {
	serverCfg := s.Config().(*Config)
	if len(cfg) > 0 {
		if err := serverv2.UnmarshalSubConfig(cfg, s.Name(), &serverCfg); err != nil {
			return fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}

	s.config = serverCfg
	return nil
}

func (s *Server[T]) Name() string {
	return ServerName
}

func (s *Server[T]) Start(context.Context) error {
	return nil
}

func (s *Server[T]) Stop(context.Context) error {
	return nil
}

func (s *Server[T]) CLICommands() serverv2.CLIConfig {
	return serverv2.CLIConfig{
		Commands: []*cobra.Command{
			s.PrunesCmd(),
			s.ExportSnapshotCmd(),
			s.DeleteSnapshotCmd(),
			s.ListSnapshotsCmd(),
			s.DumpArchiveCmd(),
			s.LoadArchiveCmd(),
			s.RestoreSnapshotCmd(s.store),
		},
	}
}

func (s *Server[T]) Config() any {
	if s.config == nil || s.config.AppDBBackend == "" {
		return DefaultConfig()
	}

	return s.config
}
