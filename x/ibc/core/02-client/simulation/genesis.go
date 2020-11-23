package simulation

import (
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
)

// GenClientGenesis returns the default client genesis state.
func GenClientGenesis(_ *rand.Rand, _ []simtypes.Account) types.GenesisState {
	return types.DefaultGenesisState()
}

// GetHistEntries randomized HistoricalEntries between 0-100.
func GetHistEntries(r *rand.Rand) uint32 {
	return uint32(r.Intn(int(types.DefaultHistoricalEntries + 1)))
}
