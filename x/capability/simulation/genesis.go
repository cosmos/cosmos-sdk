package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/capability/types"
)

// Simulation parameter constants
const index = "index"

// GenIndex returns a random global index between 1-1000
func GenIndex(r *rand.Rand) uint64 {
	return uint64(r.Int63n(1000)) + 1
}

// RandomizedGenState generates a random GenesisState for capability
func RandomizedGenState(simState *module.SimulationState) {
	var idx uint64

	simState.AppParams.GetOrGenerate(
		simState.Cdc, index, &idx, simState.Rand,
		func(r *rand.Rand) { idx = GenIndex(r) },
	)

	capabilityGenesis := types.GenesisState{Index: idx}

	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, codec.MustMarshalJSONIndent(simState.LegacyAmino, &capabilityGenesis))
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&capabilityGenesis)
}
