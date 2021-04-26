package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/feegrant/types"
)

// Simulation parameter constants
const feegrant = "feegrant"

// GenFeeGrants returns an empty slice of evidences.
func GenFeeGrants(_ *rand.Rand, _ []simtypes.Account) []types.Grant {
	return []types.Grant{}
}

// RandomizedGenState generates a random GenesisState for feegrant
func RandomizedGenState(simState *module.SimulationState) {
	var feegrants []types.Grant

	simState.AppParams.GetOrGenerate(
		simState.Cdc, feegrant, &feegrants, simState.Rand,
		func(r *rand.Rand) { feegrants = GenFeeGrants(r, simState.Accounts) },
	)
	feegrantGenesis := types.NewGenesisState(feegrants)

	bz, err := json.MarshalIndent(&feegrantGenesis, "", " ")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(feegrantGenesis)
}
