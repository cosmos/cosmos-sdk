package v048

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
)

// Migrate migrates exported state from v0.47 to a v0.48 genesis state.
func Migrate(appState types.AppMap, clientCtx client.Context) types.AppMap {
	return nil
}
