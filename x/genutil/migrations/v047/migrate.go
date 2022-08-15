package v047

import (
	"github.com/cosmos/cosmos-sdk/client"
	bankv4 "github.com/cosmos/cosmos-sdk/x/bank/migrations/v4"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
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
	return appState
}
