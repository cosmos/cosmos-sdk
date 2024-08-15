package serverv2

import (
	"encoding/json"

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
	GetGPRCMethodsToMessageMap() map[string]func() gogoproto.Message
	GetStore() any
}

// ExportedApp represents an exported app state, along with
// validators, consensus params and latest app height.
type ExportedApp struct {
	// AppState is the application state as JSON.
	AppState json.RawMessage
	// Height is the app's latest block height.
	Height int64
}
