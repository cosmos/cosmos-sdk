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

// genAuthorizationGrant returns an empty slice of authorization grants.
func genAuthorizationGrant(r *rand.Rand, accounts []simtypes.Account) []types.GrantAuthorization {
	// authorizations := make([]types.GrantAuthorization, len(accounts))

	// for i := 0; i < len(accounts); i++ {
	// 	granter, _ := simtypes.RandomAcc(r, accounts)
	// 	grantee, _ := simtypes.RandomAcc(r, accounts)
	// 	authorizations[i] = types.GrantAuthorization{
	// 		Granter:       granter.Address.String(),
	// 		Grantee:       grantee.Address.String(),
	// 		Authorization: generateRandomGrant(r),
	// 	}
	// }
	// return authorizations
	return []types.GrantAuthorization{}

}

// func generateRandomGrant(r *rand.Rand) *codectypes.Any {
// 	authorizations := make([]*codectypes.Any, 3)
// 	authorizations[0] = newAnyAuthorization(banktypes.NewSendAuthorization(sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1000)))))
// 	authorizations[1] = newAnyAuthorization(types.NewGenericAuthorization("/cosmos.gov.v1beta1.Msg/SubmitProposal"))
// 	authorizations[2] = newAnyAuthorization(types.NewGenericAuthorization("/cosmos.feegrant.v1beta1.Msg/GrantFeeAllowance"))

// 	return authorizations[r.Intn(len(authorizations))]
// }

// func newAnyAuthorization(grant exported.Authorization) *codectypes.Any {
// 	any, err := codectypes.NewAnyWithValue(grant)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return any
// }

// RandomizedGenState generates a random GenesisState for authz.
func RandomizedGenState(simState *module.SimulationState) {
	var grants []types.GrantAuthorization
	simState.AppParams.GetOrGenerate(
		simState.Cdc, authz, &grants, simState.Rand,
		func(r *rand.Rand) { grants = genAuthorizationGrant(r, simState.Accounts) },
	)

	authzGrantsGenesis := types.NewGenesisState(grants)

	bz, err := json.MarshalIndent(&authzGrantsGenesis, "", " ")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(authzGrantsGenesis)
}
