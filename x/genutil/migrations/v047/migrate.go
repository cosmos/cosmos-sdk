package v047

import (
	"github.com/cosmos/cosmos-sdk/client"
	bankv4 "github.com/cosmos/cosmos-sdk/x/bank/migrations/v4"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
	v4gov "github.com/cosmos/cosmos-sdk/x/gov/migrations/v4"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// Migrate migrates exported state from v0.46 to a v0.47 genesis state.
func Migrate(appState types.AppMap, clientCtx client.Context) types.AppMap {
	// Migrate x/bank.
	bankState := appState[banktypes.ModuleName]
	if len(bankState) > 0 {
		var oldBankState *banktypes.GenesisState
		clientCtx.Codec.MustUnmarshalJSON(bankState, oldBankState)
		newBankState := bankv4.MigrateGenState(oldBankState)
		appState[banktypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(newBankState)
	}

	if govOldState, ok := appState[v4gov.ModuleName]; ok {
		// unmarshal relative source genesis application state
		var old v1.GenesisState
		clientCtx.Codec.MustUnmarshalJSON(govOldState, &old)

		// delete deprecated x/gov genesis state
		delete(appState, v4gov.ModuleName)

		// set the x/gov genesis state with new state.
		new, err := v4gov.MigrateJSON(&old)
		if err != nil {
			panic(err)
		}
		appState[v4gov.ModuleName] = clientCtx.Codec.MustMarshalJSON(new)
	}

	return appState
}
