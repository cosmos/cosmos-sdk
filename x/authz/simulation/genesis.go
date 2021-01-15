package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/authz/types"
)

// Simulation parameter constant.
const authz = "authz"

// GenAuthorizationGrant returns an empty slice of authorization grants.
func GenAuthorizationGrant(_ *rand.Rand, _ []simtypes.Account) []types.GrantAuthorization {
	return []types.GrantAuthorization{}
}

// RandomizedGenState generates a random GenesisState for authz.
func RandomizedGenState(simState *module.SimulationState) {
	var grants []types.GrantAuthorization

	simState.AppParams.GetOrGenerate(
		simState.Cdc, authz, &grants, simState.Rand,
		func(r *rand.Rand) { grants = GenAuthorizationGrant(r, simState.Accounts) },
	)
	authzGrantsGenesis := types.NewGenesisState(grants)

	bz, err := json.MarshalIndent(&authzGrantsGenesis, "", " ")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(authzGrantsGenesis)
}
