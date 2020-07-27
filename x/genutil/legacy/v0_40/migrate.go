package v040

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	v038auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_38"
	v040auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_40"
	v038bank "github.com/cosmos/cosmos-sdk/x/bank/legacy/v0_38"
	v040bank "github.com/cosmos/cosmos-sdk/x/bank/legacy/v0_40"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
)

func Migrate(appState types.AppMap) types.AppMap {
	v038Codec := codec.New()
	cryptocodec.RegisterCrypto(v038Codec)
	v038auth.RegisterCodec(v038Codec)

	v039Codec := codec.New()
	cryptocodec.RegisterCrypto(v039Codec)
	v038auth.RegisterCodec(v039Codec)

	// remove balances from existing accounts
	if appState[v038auth.ModuleName] != nil {
		// unmarshal relative source genesis application state
		var authGenState v038auth.GenesisState
		v038Codec.MustUnmarshalJSON(appState[v038auth.ModuleName], &authGenState)

		// delete deprecated x/auth genesis state
		delete(appState, v038auth.ModuleName)

		// Migrate relative source genesis application state and marshal it into
		// the respective key.
		appState[v040auth.ModuleName] = v039Codec.MustMarshalJSON(v040auth.Migrate(authGenState))
	}

	if appState[v038bank.ModuleName] != nil {
		// unmarshal relative source genesis application state
		var bankGenState v038bank.GenesisState
		v038Codec.MustUnmarshalJSON(appState[v038bank.ModuleName], &bankGenState)

		// unmarshal x/auth genesis state to retrieve all account balances
		var authGenState v038auth.GenesisState
		v038Codec.MustUnmarshalJSON(appState[v038auth.ModuleName], &authGenState)

		// delete deprecated x/bank genesis state
		delete(appState, v038bank.ModuleName)

		// Migrate relative source genesis application state and marshal it into
		// the respective key.
		appState[v040bank.ModuleName] = v039Codec.MustMarshalJSON(v040bank.Migrate(bankGenState, authGenState))
	}

	return appState
}
