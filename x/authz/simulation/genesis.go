package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

// GenAuthorizationGrant returns an empty slice of authorization grants.
func GenAuthorizationGrant(_ *rand.Rand, _ []simtypes.Account) []authz.GrantAuthorization {
	return []authz.GrantAuthorization{}
}

// RandomizedGenState generates a random GenesisState for authz.
func RandomizedGenState(simState *module.SimulationState) {
	var grants []authz.GrantAuthorization

	simState.AppParams.GetOrGenerate(
		simState.Cdc, "authz", &grants, simState.Rand,
		func(r *rand.Rand) { grants = GenAuthorizationGrant(r, simState.Accounts) },
	)
	authzGrantsGenesis := authz.NewGenesisState(grants)

	bz, err := json.MarshalIndent(&authzGrantsGenesis, "", " ")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", authz.ModuleName, bz)
	simState.GenState[authz.ModuleName] = simState.Cdc.MustMarshalJSON(authzGrantsGenesis)
}
