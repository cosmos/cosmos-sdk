package serverv2

import (
	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/spf13/viper"

	coreapp "cosmossdk.io/core/app"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	"cosmossdk.io/server/v2/appmanager"
)

type AppCreator[T transaction.Tx] func(log.Logger, *viper.Viper) AppI[T]

type AppI[T transaction.Tx] interface {
	Name() string
	InterfaceRegistry() coreapp.InterfaceRegistry
	GetAppManager() *appmanager.AppManager[T]
	GetConsensusAuthority() string
	GetGRPCQueryDecoders() map[string]func(requestBytes []byte) (gogoproto.Message, error)
	GetStore() any
}
