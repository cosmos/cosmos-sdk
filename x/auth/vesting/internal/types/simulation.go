package types

// DONTCOVER

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// RandomGenesisAccounts returns randomly generated genesis accounts
func RandomGenesisAccounts(simState *module.SimulationState) (genesisAccs exported.GenesisAccounts) {
	for i, acc := range simState.Accounts {
		coins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(simState.InitialStake))}
		bacc := auth.NewBaseAccountWithAddress(acc.Address)
		if err := bacc.SetCoins(coins); err != nil {
			panic(err)
		}

		var gacc exported.GenesisAccount = &bacc

		// Only consider making a vesting account once the initial bonded validator
		// set is exhausted due to needing to track DelegatedVesting.
		if int64(i) > simState.NumBonded && simState.Rand.Intn(100) < 50 {
			var endTime int64

			startTime := simState.GenTimestamp.Unix()

			// Allow for some vesting accounts to vest very quickly while others very slowly.
			if simState.Rand.Intn(100) < 50 {
				endTime = int64(simulation.RandIntBetween(simState.Rand, int(startTime)+1, int(startTime+(60*60*24*30))))
			} else {
				endTime = int64(simulation.RandIntBetween(simState.Rand, int(startTime)+1, int(startTime+(60*60*12))))
			}

			if simState.Rand.Intn(100) < 50 {
				gacc = NewContinuousVestingAccount(&bacc, startTime, endTime)
			} else {
				gacc = NewDelayedVestingAccount(&bacc, endTime)
			}
		}
		genesisAccs = append(genesisAccs, gacc)
	}

	return genesisAccs
}
