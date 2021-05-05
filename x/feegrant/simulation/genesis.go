package simulation

import (
	"fmt"
	"math/rand"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/feegrant/types"
)

// Simulation parameter constants
const feegrant = "feegrant"

// genFeeGrants returns a slice of randomly generated allowances.
func genFeeGrants(r *rand.Rand, accounts []simtypes.Account) []types.Grant {
	allowances := make([]types.Grant, len(accounts)-1)
	for i := 0; i < len(accounts)-1; i++ {
		granter := accounts[i].Address
		grantee := accounts[i+1].Address
		allowances[i] = generateRandomAllowances(granter, grantee, r)
	}
	return allowances
}

func generateRandomAllowances(granter, grantee sdk.AccAddress, r *rand.Rand) types.Grant {
	allowances := make([]types.Grant, 3)
	spendLimit := sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100)))
	periodSpendLimit := sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(10)))

	basic := types.BasicAllowance{
		SpendLimit: spendLimit,
	}

	basicAllowance, err := types.NewGrant(granter, grantee, &basic)
	if err != nil {
		panic(err)
	}
	allowances[0] = basicAllowance

	periodicAllowance, err := types.NewGrant(granter, grantee, &types.PeriodicAllowance{
		Basic:            basic,
		PeriodSpendLimit: periodSpendLimit,
		Period:           time.Hour,
	})
	if err != nil {
		panic(err)
	}
	allowances[1] = periodicAllowance

	filteredAllowance, err := types.NewGrant(granter, grantee, &types.AllowedMsgAllowance{
		Allowance:       basicAllowance.GetAllowance(),
		AllowedMessages: []string{"/cosmos.gov.v1beta1.MsgSubmitProposal"},
	})
	if err != nil {
		panic(err)
	}
	allowances[2] = filteredAllowance

	return allowances[r.Intn(len(allowances))]
}

// RandomizedGenState generates a random GenesisState for feegrant
func RandomizedGenState(simState *module.SimulationState) {
	var feegrants []types.Grant

	simState.AppParams.GetOrGenerate(
		simState.Cdc, feegrant, &feegrants, simState.Rand,
		func(r *rand.Rand) { feegrants = genFeeGrants(r, simState.Accounts) },
	)

	feegrantGenesis := types.NewGenesisState(feegrants)
	bz, err := simState.Cdc.MarshalJSON(feegrantGenesis)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, bz)
	simState.GenState[types.ModuleName] = bz
}
