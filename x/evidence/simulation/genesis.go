package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/evidence/types"
)

// Simulation parameter constants
const evidence = "evidence"

// GenEvidences returns an empty slice of evidences.
func GenEvidences(_ *rand.Rand, _ []simtypes.Account) []*codectypes.Any {
	return []*codectypes.Any{}
}

// RandomizedGenState generates a random GenesisState for evidence
func RandomizedGenState(simState *module.SimulationState) {
	var ev []*codectypes.Any

	simState.AppParams.GetOrGenerate(
		simState.Cdc, evidence, &ev, simState.Rand,
		func(r *rand.Rand) { ev = GenEvidences(r, simState.Accounts) },
	)

	evidenceGenesis := types.GenesisState{Evidence: ev}

	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, codec.MustMarshalJSONIndent(simState.Cdc, evidenceGenesis))
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(evidenceGenesis)
}
