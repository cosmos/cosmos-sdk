package serverv2

import (
	"github.com/spf13/viper"

	servercore "cosmossdk.io/core/server"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
)

type AppCreator[AppT servercore.AppI[T], T transaction.Tx] func(log.Logger, *viper.Viper) AppT
