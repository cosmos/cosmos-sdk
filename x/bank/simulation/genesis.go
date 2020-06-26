package simulation

// DONTCOVER

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// Simulation parameter constants
const (
	SendEnabled = "send_enabled_params"
)

// GenSendEnabledParams randomized SendEnabledParams
func GenSendEnabledParams(r *rand.Rand) types.SendEnabledParams {
	var sendParams types.SendEnabledParams

	if r.Int63n(101) <= 95 { // 95% chance of transfers being enabled
		sendParams = append(sendParams, types.DefaultSendEnabledParam())
	}

	if r.Int63n(101) <= 50 { // half the time add an additional denom specific record
		sendParams = append(sendParams, types.NewSendEnabledParam(
			sdk.DefaultBondDenom,
			r.Int63n(101) <= 95))
	}

	return sendParams
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
		simState.Cdc, SendEnabled, &sendEnabledParams, simState.Rand,
		func(r *rand.Rand) { sendEnabledParams = GenSendEnabledParams(r) },
	)

	numAccs := int64(len(simState.Accounts))
	totalSupply := sdk.NewInt(simState.InitialStake * (numAccs + simState.NumBonded))
	supply := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, totalSupply))

	bankGenesis := types.NewGenesisState(sendEnabledParams, RandomGenesisBalances(simState), supply)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(bankGenesis)
}
