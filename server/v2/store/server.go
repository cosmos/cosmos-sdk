package store

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	serverv2 "cosmossdk.io/server/v2"
)

type StoreComponent struct{}

func New() serverv2.ServerComponent[transaction.Tx] {
	return StoreComponent{}
}

func (s StoreComponent) Init(appI serverv2.AppI[transaction.Tx], v *viper.Viper, logger log.Logger) (serverv2.ServerComponent[transaction.Tx], error) {
	return s, nil
}

func (s StoreComponent) Name() string {
	return "store"
}

func (s StoreComponent) Start(ctx context.Context) error {
	return nil
}

func (s StoreComponent) Stop(ctx context.Context) error {
	return nil
}

func (s StoreComponent) CLICommands(appCreator serverv2.AppCreator[transaction.Tx]) serverv2.CLIConfig {
	return serverv2.CLIConfig{
		Commands: []*cobra.Command{
			s.PrunesCmd(appCreator),
		},
	}
}
