package store

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	servercore "cosmossdk.io/core/server"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	// serverv2 "cosmossdk.io/server/v2"
)

type StoreComponent struct {
	config *Config
}

func New() *StoreComponent {
	return &StoreComponent{}
}

func (s *StoreComponent) Init(appI servercore.AppI[transaction.Tx], v *viper.Viper, logger log.Logger) error {
	cfg := DefaultConfig()
	if v != nil {
		if err := v.Sub(s.Name()).Unmarshal(&cfg); err != nil {
			return fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}
	s.config = cfg
	return nil
}

func (s *StoreComponent) Name() string {
	return "store"
}

func (s *StoreComponent) Start(ctx context.Context) error {
	return nil
}

func (s *StoreComponent) Stop(ctx context.Context) error {
	return nil
}

func (s *StoreComponent) CLICommands(appCreator servercore.AppCreator[transaction.Tx]) servercore.CLIConfig {
	return servercore.CLIConfig{
		Commands: []*cobra.Command{
			s.PrunesCmd(appCreator),
		},
	}
}

func (g *StoreComponent) Config() any {
	if g.config == nil || g.config == (&Config{}) {
		return DefaultConfig()
	}

	return g.config
}
