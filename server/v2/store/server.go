package store

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	serverv2 "cosmossdk.io/server/v2"
)

// StoreComponent manages store config
// and contains prune & snapshot commands
type StoreComponent[T transaction.Tx] struct {
	config *Config
}

func New[T transaction.Tx]() *StoreComponent[T] {
	return &StoreComponent[T]{}
}

func (s *StoreComponent[T]) Init(appI serverv2.AppI[T], v *viper.Viper, logger log.Logger) error {
	cfg := DefaultConfig()
	if v != nil {
		if err := v.Sub(s.Name()).Unmarshal(&cfg); err != nil {
			return fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}
	s.config = cfg
	return nil
}

func (s *StoreComponent[T]) Name() string {
	return "store"
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
		},
	}
}

func (g *StoreComponent[T]) Config() any {
	if g.config == nil || g.config == (&Config{}) {
		return DefaultConfig()
	}

	return g.config
}
