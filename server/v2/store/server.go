package store

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	serverv2 "cosmossdk.io/server/v2"
)

var (
	_ serverv2.ServerComponent[transaction.Tx] = (*StoreComponent[transaction.Tx])(nil)
	_ serverv2.HasConfig                       = (*StoreComponent[transaction.Tx])(nil)
	_ serverv2.HasCLICommands                  = (*StoreComponent[transaction.Tx])(nil)
)

const ServerName = "store"

// StoreComponent manages store config
// and contains prune & snapshot commands
type StoreComponent[T transaction.Tx] struct {
	config *Config
	// saving appCreator for only RestoreSnapshotCmd
	appCreator serverv2.AppCreator[T]
}

func New[T transaction.Tx](appCreator serverv2.AppCreator[T]) *StoreComponent[T] {
	return &StoreComponent[T]{appCreator: appCreator}
}

func (s *StoreComponent[T]) Init(appI serverv2.AppI[T], cfg map[string]any, logger log.Logger) error {
	serverCfg := DefaultConfig()
	if len(cfg) > 0 {
		if err := serverv2.UnmarshalSubConfig(cfg, s.Name(), &serverCfg); err != nil {
			return fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}
	s.config = serverCfg
	return nil
}

func (s *StoreComponent[T]) Name() string {
	return ServerName
}

func (s *StoreComponent[T]) Start(ctx context.Context) error {
	return nil
}

func (s *StoreComponent[T]) Stop(ctx context.Context) error {
	return nil
}

func (s *StoreComponent[T]) CLICommands() serverv2.CLIConfig {
	return serverv2.CLIConfig{
		Commands: []*cobra.Command{
			s.PrunesCmd(),
			s.ExportSnapshotCmd(),
			s.DeleteSnapshotCmd(),
			s.ListSnapshotsCmd(),
			s.DumpArchiveCmd(),
			s.LoadArchiveCmd(),
			s.RestoreSnapshotCmd(s.appCreator),
		},
	}
}

func (g *StoreComponent[T]) Config() any {
	if g.config == nil || g.config == (&Config{}) {
		return DefaultConfig()
	}

	return g.config
}
