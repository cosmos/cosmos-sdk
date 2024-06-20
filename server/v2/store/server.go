package store

import (
	"context"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	serverv2 "cosmossdk.io/server/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type StoreServer struct {}

func (s StoreServer) Init(appI serverv2.AppI[transaction.Tx], v *viper.Viper, logger log.Logger) (serverv2.ServerComponent[transaction.Tx], error) {
	return s, nil
}

func (s StoreServer) Name() string {
	return "store"
}

func (s StoreServer) Start(ctx context.Context) error {
	return nil
}

func (s StoreServer) Stop(ctx context.Context) error {
	return nil
}

func (s StoreServer) CLICommands(appCreator serverv2.AppCreator[transaction.Tx]) serverv2.CLIConfig {
	return serverv2.CLIConfig{
		Commands: []*cobra.Command{
			s.PrunesCmd(appCreator),
		},
	}
}


