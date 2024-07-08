package store

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	servercore "cosmossdk.io/core/server"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/core/log"
)

type StoreComponent[AppT servercore.AppI[T], T transaction.Tx] struct {
	config *Config
}

func New[AppT servercore.AppI[T], T transaction.Tx]() *StoreComponent[AppT, T] {
	return &StoreComponent[AppT, T]{}
}

func (s *StoreComponent[AppT, T]) Init(appI AppT, v *viper.Viper, logger log.Logger) error {
	cfg := DefaultConfig()
	if v != nil {
		if err := v.Sub(s.Name()).Unmarshal(&cfg); err != nil {
			return fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}
	s.config = cfg
	return nil
}

func (s *StoreComponent[AppT, T]) Name() string {
	return "store"
}

func (s *StoreComponent[AppT, T]) Start(ctx context.Context) error {
	return nil
}

func (s *StoreComponent[AppT, T]) Stop(ctx context.Context) error {
	return nil
}

func (s *StoreComponent[AppT, T]) GetCommands() []*cobra.Command {
	return []*cobra.Command{
		s.PrunesCmd(),
	}
}

func (s *StoreComponent[AppT, T]) GetTxs() []*cobra.Command {
	return nil
}

func (s *StoreComponent[AppT, T]) GetQueries() []*cobra.Command {
	return nil
}

func (g *StoreComponent[AppT, T]) Config() any {
	if g.config == nil || g.config == (&Config{}) {
		return DefaultConfig()
	}

	return g.config
}
