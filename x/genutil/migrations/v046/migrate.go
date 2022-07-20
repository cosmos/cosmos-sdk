package v046

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
	v043gov "github.com/cosmos/cosmos-sdk/x/gov/migrations/v043"
	v046gov "github.com/cosmos/cosmos-sdk/x/gov/migrations/v046"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingv2 "github.com/cosmos/cosmos-sdk/x/staking/migrations/v2"
	stakingv3 "github.com/cosmos/cosmos-sdk/x/staking/migrations/v3"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Migrate migrates exported state from v0.43 to a v0.46 genesis state.
func Migrate(appState types.AppMap, clientCtx client.Context) types.AppMap {
	// Migrate x/gov.
	if appState[v043gov.ModuleName] != nil {
		// unmarshal relative source genesis application state
		var old govv1beta1.GenesisState
		clientCtx.Codec.MustUnmarshalJSON(appState[v043gov.ModuleName], &old)

		// delete deprecated x/gov genesis state
		delete(appState, v043gov.ModuleName)

		// Migrate relative source genesis application state and marshal it into
		// the respective key.
		new, err := v046gov.MigrateJSON(&old)
		if err != nil {
			panic(err)
		}
		appState[v046gov.ModuleName] = clientCtx.Codec.MustMarshalJSON(new)
	}

	// Migrate x/staking.
	if appState[stakingv2.ModuleName] != nil {
		// unmarshal relative source genesis application state
		var old stakingtypes.GenesisState
		clientCtx.Codec.MustUnmarshalJSON(appState[stakingv2.ModuleName], &old)

		// delete deprecated x/staking genesis state
		delete(appState, stakingv2.ModuleName)

		// Migrate relative source genesis application state and marshal it into
		// the respective key.
		new, err := stakingv3.MigrateJSON(old)
		if err != nil {
			panic(err)
		}
		appState[stakingv3.ModuleName] = clientCtx.Codec.MustMarshalJSON(&new)
	}

	return appState
}
