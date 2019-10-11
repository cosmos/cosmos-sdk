package v038

import (
	"github.com/cosmos/cosmos-sdk/codec"
	v036auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_36"
	v038auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_38"
	v036genaccounts "github.com/cosmos/cosmos-sdk/x/genaccounts/legacy/v0_36"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	v036staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v0_36"
	v038staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v0_38"
)

// Migrate migrates exported state from v0.34 to a v0.36 genesis state.
func Migrate(appState genutil.AppMap) genutil.AppMap {
	v036Codec := codec.New()
	codec.RegisterCrypto(v036Codec)

	v038Codec := codec.New()
	codec.RegisterCrypto(v038Codec)
	v038auth.RegisterCodec(v038Codec)

	if appState[v036genaccounts.ModuleName] != nil {
		// unmarshal relative source genesis application state
		var authGenState v036auth.GenesisState
		v036Codec.MustUnmarshalJSON(appState[v036auth.ModuleName], &authGenState)

		var genAccountsGenState v036genaccounts.GenesisState
		v036Codec.MustUnmarshalJSON(appState[v036genaccounts.ModuleName], &genAccountsGenState)

		// Migrate relative source genesis application state and marshal it into
		// the respective key.
		appState[v038auth.ModuleName] = v038Codec.MustMarshalJSON(
			v038auth.Migrate(authGenState, genAccountsGenState),
		)

		// delete deprecated genaccounts genesis state
		delete(appState, v036genaccounts.ModuleName)
	}

	// migrate staking state
	if appState[v036staking.ModuleName] != nil {
		var stakingGenState v036staking.GenesisState
		v036Codec.MustUnmarshalJSON(appState[v036staking.ModuleName], &stakingGenState)

		delete(appState, v036staking.ModuleName) // delete old key in case the name changed
		appState[v038staking.ModuleName] = v038Codec.MustMarshalJSON(v038staking.Migrate(stakingGenState))
	}

	return appState
}
