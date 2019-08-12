package simulation

// DONTCOVER

import (
	"encoding/json"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/genaccounts/internal/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// RandomizedGenState generates a random GenesisState for the genesis accounts
func RandomizedGenState(
	cdc *codec.Codec, r *rand.Rand, genesisState map[string]json.RawMessage,
	accs []simulation.Account, genesisTimestamp time.Time, amount, numInitiallyBonded int64,
) {

	var genesisAccounts []types.GenesisAccount

	// randomly generate some genesis accounts
	for i, acc := range accs {
		coins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(amount))}
		bacc := authtypes.NewBaseAccountWithAddress(acc.Address)
		if err := bacc.SetCoins(coins); err != nil {
			panic(err)
		}

		var gacc types.GenesisAccount

		// Only consider making a vesting account once the initial bonded validator
		// set is exhausted due to needing to track DelegatedVesting.
		if int64(i) > numInitiallyBonded && r.Intn(100) < 50 {
			var (
				vacc    authexported.VestingAccount
				endTime int64
			)

			startTime := genesisTimestamp.Unix()

			// Allow for some vesting accounts to vest very quickly while others very slowly.
			if r.Intn(100) < 50 {
				endTime = int64(simulation.RandIntBetween(r, int(startTime), int(startTime+(60*60*24*30))))
			} else {
				endTime = int64(simulation.RandIntBetween(r, int(startTime), int(startTime+(60*60*12))))
			}

			if startTime == endTime {
				endTime++
			}

			if r.Intn(100) < 50 {
				vacc = authtypes.NewContinuousVestingAccount(&bacc, startTime, endTime)
			} else {
				vacc = authtypes.NewDelayedVestingAccount(&bacc, endTime)
			}

			var err error
			gacc, err = types.NewGenesisAccountI(vacc)
			if err != nil {
				panic(err)
			}
		} else {
			gacc = types.NewGenesisAccount(&bacc)
		}

		genesisAccounts = append(genesisAccounts, gacc)
	}

	genesisState[types.ModuleName] = cdc.MustMarshalJSON(genesisAccounts)
}
