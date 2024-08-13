package serverv2

import (
	"encoding/json"
	dbm "github.com/cosmos/cosmos-db"
	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/spf13/viper"
	"io"

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

// AppExporter is a function that dumps all app state to
// JSON-serializable structure and returns the current validator set.
type AppExporter func(
	logger log.Logger,
	db dbm.DB,
	traceWriter io.Writer,
	height int64,
	forZeroHeight bool,
	jailAllowedAddrs []string,
	opts AppOptions,
	modulesToExport []string,
) (ExportedApp, error)

// AppOptions defines an interface that is passed into an application
// constructor, typically used to set BaseApp options that are either supplied
// via config file or through CLI arguments/flags. The underlying implementation
// is defined by the server package and is typically implemented via a Viper
// literal defined on the server Context. Note, casting Get calls may not yield
// the expected types and could result in type assertion errors. It is recommend
// to either use the cast package or perform manual conversion for safety.
type AppOptions interface {
	Get(string) interface{}
}
