package v2

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ExportedApp represents an exported app state, along with
// validators, consensus params and latest app height.
type ExportedApp struct {
	// AppState is the application state as JSON.
	AppState json.RawMessage
	// Height is the app's latest block height.
	Height int64
	// Validators is the exported validator set.
	Validators []sdk.GenesisValidator
}
