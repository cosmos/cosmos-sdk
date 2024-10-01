package serverv2

import (
	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/spf13/viper"

	"cosmossdk.io/core/server"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	"cosmossdk.io/server/v2/appmanager"
)

type AppCreator[T transaction.Tx] func(log.Logger, *viper.Viper) AppI[T]

type AppI[T transaction.Tx] interface {
	Name() string
	InterfaceRegistry() server.InterfaceRegistry
	GetAppManager() *appmanager.AppManager[T]
	GetGPRCMethodsToMessageMap() map[string]func() gogoproto.Message
	GetStore() any
}
