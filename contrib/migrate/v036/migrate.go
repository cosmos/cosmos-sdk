package v036

import (
	"github.com/cosmos/cosmos-sdk/codec"
	extypes "github.com/cosmos/cosmos-sdk/contrib/migrate/types"
	v034gov "github.com/cosmos/cosmos-sdk/x/gov/legacy/v034"
	v036gov "github.com/cosmos/cosmos-sdk/x/gov/legacy/v036"
)

// Migrate migrates exported state from v0.34 to a v0.36 genesis state.
func Migrate(appState extypes.AppMap, cdc *codec.Codec) extypes.AppMap {
	v034Codec := codec.New()
	codec.RegisterCrypto(v034Codec)
	v036Codec := codec.New()
	codec.RegisterCrypto(v036Codec)

	if appState[v034gov.ModuleName] != nil {
		var govState v034gov.GenesisState
		v034gov.RegisterCodec(v034Codec)
		v034Codec.MustUnmarshalJSON(appState[v034gov.ModuleName], &govState)
		v036gov.RegisterCodec(v036Codec)
		delete(appState, v034gov.ModuleName) // Drop old key, in case it changed name
		appState[v036gov.ModuleName] = v036Codec.MustMarshalJSON(v036gov.MigrateGovernance(govState))
	}
	return appState
}
