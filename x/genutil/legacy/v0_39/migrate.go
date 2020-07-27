package v039

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	v038auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_38"
	v039auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_39"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
)

// Migrate accepts exported x/auth genesis state from v0.38 and migrates it to
// v0.39 x/auth genesis state. The migration includes:
//
// - Public key encoding being changed from bech32 to Amino
func Migrate(appState types.AppMap) types.AppMap {
	v038Codec := codec.New()
	cryptocodec.RegisterCrypto(v038Codec)
	v038auth.RegisterCodec(v038Codec)

	v039Codec := codec.New()
	cryptocodec.RegisterCrypto(v039Codec)
	v039auth.RegisterCodec(v039Codec)

	// migrate x/auth state (JSON serialization only)
	if appState[v038auth.ModuleName] != nil {
		var authGenState v038auth.GenesisState
		v038Codec.MustUnmarshalJSON(appState[v038auth.ModuleName], &authGenState)

		delete(appState, v038auth.ModuleName) // delete old key in case the name changed
		appState[v039auth.ModuleName] = v039Codec.MustMarshalJSON(authGenState)
	}

	return appState
}
