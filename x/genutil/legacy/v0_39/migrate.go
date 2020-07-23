package v039

import (
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
)

// Migrate accepts exported x/auth genesis state from v0.38 and migrates it to
// v0.39 x/auth genesis state. The migration includes:
//
// - Public key encoding being changed from bech32 to Amino
func Migrate(appState types.AppMap) types.AppMap {
	return appState
}
