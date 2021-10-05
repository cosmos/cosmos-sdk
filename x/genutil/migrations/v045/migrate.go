package v045

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"

	v043gov "github.com/cosmos/cosmos-sdk/x/gov/migrations/v043"
	v045gov "github.com/cosmos/cosmos-sdk/x/gov/migrations/v045"
)

func Migrate(appState types.AppMap, clientCtx client.Context) types.AppMap {
	// Migrate x/gov proposals
	if appState[v043gov.ModuleName] != nil {
		// unmarshal relative source genesis application state
		var oldGovState v043gov.GenesisState
		clientCtx.Codec.MustUnmarshalJSON(appState[v043gov.ModuleName], &oldGovState)

		// delete deprecated x/gov genesis state
		delete(appState, v043gov.ModuleName)

		// Migrate relative source genesis application state and marshal it into
		// the respective key.
		appState[v043gov.ModuleName] = clientCtx.Codec.MustMarshalJSON(v045gov.Migrate(&oldGovState))
	}
}