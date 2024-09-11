package v2

import (
	"encoding/json"

	"cosmossdk.io/log"
)

// AppExporter is a function that dumps all app state to
// JSON-serializable structure and returns the current validator set.
type AppExporter func(
	logger log.Logger,
	height int64,
	jailAllowedAddrs []string,
) (ExportedApp, error)

// ExportedApp represents an exported app state, along with
// validators, consensus params and latest app height.
type ExportedApp struct {
	// AppState is the application state as JSON.
	AppState json.RawMessage
	// Height is the app's latest block height.
	Height int64
}
