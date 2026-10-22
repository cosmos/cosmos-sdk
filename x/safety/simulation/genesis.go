package simulation

import (
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/safety/types"
)

// RandomizedGenState generates a random GenesisState for x/safety. The
// authority is fixed; the operator is a random simulation account or empty,
// and the disabled list starts empty so that other modules' operations run.
func RandomizedGenState(simState *module.SimulationState, authority string) {
	params := types.DefaultParams(authority)
	if len(simState.Accounts) > 0 && simState.Rand.Intn(2) == 0 {
		params.Operator = simState.Accounts[simState.Rand.Intn(len(simState.Accounts))].Address.String()
	}
	if simState.Rand.Intn(2) == 0 {
		params.Scope = []string{"/cosmos.bank.*", "/cosmos.staking.*"}
	}

	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(types.NewGenesisState(params, nil))
}
