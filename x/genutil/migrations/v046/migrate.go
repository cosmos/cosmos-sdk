package v046

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
	v043gov "github.com/cosmos/cosmos-sdk/x/gov/migrations/v043"
	v046gov "github.com/cosmos/cosmos-sdk/x/gov/migrations/v046"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// Migrate migrates exported state from v0.43 to a v0.46 genesis state.
func Migrate(appState types.AppMap, clientCtx client.Context) types.AppMap {
	// Migrate x/gov.
	if appState[v043gov.ModuleName] != nil {
		// unmarshal relative source genesis application state
		var oldGovState gov.GenesisState
		clientCtx.Codec.MustUnmarshalJSON(appState[v043gov.ModuleName], &oldGovState)

		// delete deprecated x/gov genesis state
		delete(appState, v043gov.ModuleName)

		// Migrate relative source genesis application state and marshal it into
		// the respective key.
		newGovState, err := v046gov.MigrateJSON(&oldGovState)
		if err != nil {
			panic(err)
		}
		appState[v046gov.ModuleName] = clientCtx.Codec.MustMarshalJSON(newGovState)
	}

	return appState
}
