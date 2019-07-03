package v036

import (
	"github.com/cosmos/cosmos-sdk/codec"
	v034distr "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v0_34"
	v036distr "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v0_36"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	v034gov "github.com/cosmos/cosmos-sdk/x/gov/legacy/v034"
	v036gov "github.com/cosmos/cosmos-sdk/x/gov/legacy/v036"
)

// Migrate migrates exported state from v0.34 to a v0.36 genesis state.
func Migrate(appState genutil.AppMap) genutil.AppMap {
	v034Codec := codec.New()
	codec.RegisterCrypto(v034Codec)

	v036Codec := codec.New()
	codec.RegisterCrypto(v036Codec)

	// migrate genesis state
	if appState[v034gov.ModuleName] != nil {
		v034gov.RegisterCodec(v034Codec)
		v036gov.RegisterCodec(v036Codec)

		var govState v034gov.GenesisState
		v034Codec.MustUnmarshalJSON(appState[v034gov.ModuleName], &govState)

		delete(appState, v034gov.ModuleName) // delete old key in case the name changed
		appState[v036gov.ModuleName] = v036Codec.MustMarshalJSON(v036gov.MigrateGovernance(govState))
	}

	// migrate distribution state
	if appState[v034distr.ModuleName] != nil {
		var slashingGenState v034distr.GenesisState
		v034Codec.MustUnmarshalJSON(appState[v034distr.ModuleName], &slashingGenState)

		delete(appState, v034distr.ModuleName) // delete old key in case the name changed
		appState[v036distr.ModuleName] = v036Codec.MustMarshalJSON(v036distr.Migrate(slashingGenState))
	}

	return appState
}
