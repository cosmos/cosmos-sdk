package v040

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	v039auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_39"
	v040auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_40"
	v038bank "github.com/cosmos/cosmos-sdk/x/bank/legacy/v0_38"
	v040bank "github.com/cosmos/cosmos-sdk/x/bank/legacy/v0_40"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
)

// Migrate migrates exported state from v0.39 to a v0.40 genesis state.
func Migrate(appState types.AppMap) types.AppMap {
	v039Codec := codec.New()
	cryptocodec.RegisterCrypto(v039Codec)
	v039auth.RegisterCodec(v039Codec)

	v040Codec := codec.New()
	cryptocodec.RegisterCrypto(v040Codec)
	v039auth.RegisterCodec(v040Codec)

	// remove balances from existing accounts
	if appState[v039auth.ModuleName] != nil {
		// unmarshal relative source genesis application state
		var authGenState v039auth.GenesisState
		v039Codec.MustUnmarshalJSON(appState[v039auth.ModuleName], &authGenState)

		// delete deprecated x/auth genesis state
		delete(appState, v039auth.ModuleName)

		// Migrate relative source genesis application state and marshal it into
		// the respective key.
		appState[v040auth.ModuleName] = v040Codec.MustMarshalJSON(v040auth.Migrate(authGenState))
	}

	if appState[v038bank.ModuleName] != nil {
		// unmarshal relative source genesis application state
		var bankGenState v038bank.GenesisState
		v039Codec.MustUnmarshalJSON(appState[v038bank.ModuleName], &bankGenState)

		// unmarshal x/auth genesis state to retrieve all account balances
		var authGenState v039auth.GenesisState
		v039Codec.MustUnmarshalJSON(appState[v039auth.ModuleName], &authGenState)

		// delete deprecated x/bank genesis state
		delete(appState, v038bank.ModuleName)

		// Migrate relative source genesis application state and marshal it into
		// the respective key.
		appState[v040bank.ModuleName] = v040Codec.MustMarshalJSON(v040bank.Migrate(bankGenState, authGenState))
	}

	return appState
}
