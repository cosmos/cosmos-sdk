package v038

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	v036auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v036"
	v038auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v038"
	v036distr "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v036"
	v038distr "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v038"
	v036genaccounts "github.com/cosmos/cosmos-sdk/x/genaccounts/legacy/v036"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
	v036gov "github.com/cosmos/cosmos-sdk/x/gov/legacy/v036"
	v036params "github.com/cosmos/cosmos-sdk/x/params/legacy/v036"
	v036staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v036"
	v038staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v038"
	v038upgrade "github.com/cosmos/cosmos-sdk/x/upgrade/legacy/v038"
)

// Migrate migrates exported state from v0.36/v0.37 to a v0.38 genesis state.
func Migrate(appState types.AppMap, _ client.Context) types.AppMap {
	v036Codec := codec.NewLegacyAmino()
	cryptocodec.RegisterCrypto(v036Codec)
	v036gov.RegisterLegacyAminoCodec(v036Codec)
	v036distr.RegisterLegacyAminoCodec(v036Codec)
	v036params.RegisterLegacyAminoCodec(v036Codec)

	v038Codec := codec.NewLegacyAmino()
	v038auth.RegisterLegacyAminoCodec(v038Codec)
	v036gov.RegisterLegacyAminoCodec(v038Codec)
	v036distr.RegisterLegacyAminoCodec(v038Codec)
	v036params.RegisterLegacyAminoCodec(v038Codec)
	v038upgrade.RegisterLegacyAminoCodec(v038Codec)

	if appState[v036genaccounts.ModuleName] != nil {
		// unmarshal relative source genesis application state
		var authGenState v036auth.GenesisState
		v036Codec.MustUnmarshalJSON(appState[v036auth.ModuleName], &authGenState)

		var genAccountsGenState v036genaccounts.GenesisState
		v036Codec.MustUnmarshalJSON(appState[v036genaccounts.ModuleName], &genAccountsGenState)

		// delete deprecated genaccounts genesis state
		delete(appState, v036genaccounts.ModuleName)

		// Migrate relative source genesis application state and marshal it into
		// the respective key.
		appState[v038auth.ModuleName] = v038Codec.MustMarshalJSON(v038auth.Migrate(authGenState, genAccountsGenState))
	}

	// migrate staking state
	if appState[v036staking.ModuleName] != nil {
		var stakingGenState v036staking.GenesisState
		v036Codec.MustUnmarshalJSON(appState[v036staking.ModuleName], &stakingGenState)

		delete(appState, v036staking.ModuleName) // delete old key in case the name changed
		appState[v038staking.ModuleName] = v038Codec.MustMarshalJSON(v038staking.Migrate(stakingGenState))
	}

	// migrate distribution state
	if appState[v036distr.ModuleName] != nil {
		var distrGenState v036distr.GenesisState
		v036Codec.MustUnmarshalJSON(appState[v036distr.ModuleName], &distrGenState)

		delete(appState, v036distr.ModuleName) // delete old key in case the name changed
		appState[v038distr.ModuleName] = v038Codec.MustMarshalJSON(v038distr.Migrate(distrGenState))
	}

	return appState
}
