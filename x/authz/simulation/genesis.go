package simulation

import (
	"fmt"
	"math/rand"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/authz/exported"
	"github.com/cosmos/cosmos-sdk/x/authz/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// Simulation parameter constant.
const authz = "authz"

// genAuthorizationGrant returns an empty slice of authorization grants.
func genAuthorizationGrant(r *rand.Rand, accounts []simtypes.Account) []types.GrantAuthorization {
	authorizations := make([]types.GrantAuthorization, len(accounts)-1)
	for i := 0; i < len(accounts)-1; i++ {
		granter := accounts[i]
		grantee := accounts[i+1]
		authorizations[i] = types.GrantAuthorization{
			Granter:       granter.Address.String(),
			Grantee:       grantee.Address.String(),
			Authorization: generateRandomGrant(r),
		}
	}

	return authorizations
}

func generateRandomGrant(r *rand.Rand) *codectypes.Any {
	authorizations := make([]*codectypes.Any, 1)
	authorizations[0] = newAnyAuthorization(banktypes.NewSendAuthorization(sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1000)))))
	// authorizations[1] = newAnyAuthorization(types.NewGenericAuthorization("/cosmos.gov.v1beta1.Msg/SubmitProposal"))

	return authorizations[r.Intn(len(authorizations))]
}

func newAnyAuthorization(grant exported.Authorization) *codectypes.Any {
	any, err := codectypes.NewAnyWithValue(grant)
	if err != nil {
		panic(err)
	}

	return any
}

// RandomizedGenState generates a random GenesisState for authz.
func RandomizedGenState(simState *module.SimulationState) {
	var grants []types.GrantAuthorization
	simState.AppParams.GetOrGenerate(
		simState.Cdc, authz, &grants, simState.Rand,
		func(r *rand.Rand) { grants = genAuthorizationGrant(r, simState.Accounts) },
	)

	authzGrantsGenesis := types.NewGenesisState(grants)

	bz, err := simState.Cdc.MarshalJSON(authzGrantsGenesis)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(authzGrantsGenesis)
}
