package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// Simulation parameter constants
const (
	KeySendEnabled = "SendEnabled"
)

// RandomGenesisSendParams randomized Parameters for the bank module
func RandomGenesisSendParams(r *rand.Rand) types.SendEnabledParams {
	params := types.DefaultParams()

	// 10% chance of transfers being disabled or P(a) = 0.9 for success
	if r.Int63n(101) <= 10 {
		params = params.SetSendEnabledParam("", false)
	}

	// 50% of the time add an additional denom specific record (P(b) = 0.475 = 0.5 * 0.95)
	if r.Int63n(101) <= 50 {
		// set send enabled 95% of the time
		bondEnabled := r.Int63n(101) <= 95
		params = params.SetSendEnabledParam(
			sdk.DefaultBondDenom,
			bondEnabled)
	}

	// overall probability of enabled for bond denom is 94.75% (P(a)+P(b) - P(a)*P(b))
	return params.SendEnabled
}

// RandomGenesisAccounts returns a slice of account balances. Each account has
// a balance of simState.InitialStake for sdk.DefaultBondDenom.
func RandomGenesisBalances(simState *module.SimulationState) []types.Balance {
	genesisBalances := []types.Balance{}

	for _, acc := range simState.Accounts {
		genesisBalances = append(genesisBalances, types.Balance{
			Address: acc.Address,
			Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(simState.InitialStake))),
		})
	}

	return genesisBalances
}

// RandomizedGenState generates a random GenesisState for bank
func RandomizedGenState(simState *module.SimulationState) {
	var sendEnabledParams types.SendEnabledParams
	simState.AppParams.GetOrGenerate(
		simState.Cdc, KeySendEnabled, &sendEnabledParams, simState.Rand,
		func(r *rand.Rand) { sendEnabledParams = RandomGenesisSendParams(r) },
	)

	numAccs := int64(len(simState.Accounts))
	totalSupply := sdk.NewInt(simState.InitialStake * (numAccs + simState.NumBonded))
	supply := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, totalSupply))

	bankGenesis := types.GenesisState{
		Params: types.Params{
			SendEnabled: sendEnabledParams,
		},
		Balances: RandomGenesisBalances(simState),
		Supply:   supply,
	}

	fmt.Printf("Selected randomly generated bank parameters:\n%s\n", codec.MustMarshalJSONIndent(simState.Cdc, bankGenesis.Params))
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(bankGenesis)
}
