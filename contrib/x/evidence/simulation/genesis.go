package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/contrib/x/evidence/exported"
	types2 "github.com/cosmos/cosmos-sdk/contrib/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

// Simulation parameter constants
const evidence = "evidence"

// GenEvidences returns an empty slice of evidences.
func GenEvidences(_ *rand.Rand, _ []simtypes.Account) []exported.Evidence {
	return []exported.Evidence{}
}

// RandomizedGenState generates a random GenesisState for evidence
func RandomizedGenState(simState *module.SimulationState) {
	var ev []exported.Evidence

	simState.AppParams.GetOrGenerate(evidence, &ev, simState.Rand, func(r *rand.Rand) { ev = GenEvidences(r, simState.Accounts) })

	evidenceGenesis := types2.NewGenesisState(ev)
	simState.GenState[types2.ModuleName] = simState.Cdc.MustMarshalJSON(evidenceGenesis)
}
