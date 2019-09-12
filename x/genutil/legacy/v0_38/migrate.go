package v038

import (
	"github.com/cosmos/cosmos-sdk/codec"
	v036auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_36"
	v038auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_38"
	v036genAccounts "github.com/cosmos/cosmos-sdk/x/genaccounts/legacy/v0_36"
	"github.com/cosmos/cosmos-sdk/x/genutil"
)

// Migrate migrates exported state from v0.34 to a v0.36 genesis state.
func Migrate(appState genutil.AppMap) genutil.AppMap {
	// migrate genesis accounts state

	v038Codec := codec.New()
	codec.RegisterCrypto(v038Codec)

	if appState[v036genAccounts.ModuleName] != nil {

		var authGenState v036auth.GenesisState
		v038Codec.MustUnmarshalJSON(appState[v036auth.ModuleName], &authGenState)

		appState[v038auth.ModuleName] = v038Codec.MustMarshalJSON(
			v038auth.Migrate(authGenState, appState[v036genAccounts.ModuleName]),
		)

		delete(appState, v036genAccounts.ModuleName)
	}

	return appState
}
