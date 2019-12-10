package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// Simulation parameter constants
const (
	MaxMemoChars           = "max_memo_characters"
	TxSigLimit             = "tx_sig_limit"
	TxSizeCostPerByte      = "tx_size_cost_per_byte"
	SigVerifyCostED25519   = "sig_verify_cost_ed25519"
	SigVerifyCostSECP256K1 = "sig_verify_cost_secp256k1"
)

// GenMaxMemoChars randomized MaxMemoChars
func GenMaxMemoChars(r *rand.Rand) uint64 {
	return uint64(simulation.RandIntBetween(r, 100, 200))
}

// GenTxSigLimit randomized TxSigLimit
// make sure that sigLimit is always high
// so that arbitrarily simulated messages from other
// modules can still create valid transactions
func GenTxSigLimit(r *rand.Rand) uint64 {
	return uint64(r.Intn(7) + 5)
}

// GenTxSizeCostPerByte randomized TxSizeCostPerByte
func GenTxSizeCostPerByte(r *rand.Rand) uint64 {
	return uint64(simulation.RandIntBetween(r, 5, 15))
}

// GenSigVerifyCostED25519 randomized SigVerifyCostED25519
func GenSigVerifyCostED25519(r *rand.Rand) uint64 {
	return uint64(simulation.RandIntBetween(r, 500, 1000))
}

// GenSigVerifyCostSECP256K1 randomized SigVerifyCostSECP256K1
func GenSigVerifyCostSECP256K1(r *rand.Rand) uint64 {
	return uint64(simulation.RandIntBetween(r, 500, 1000))
}

// RandomizedGenState generates a random GenesisState for auth
func RandomizedGenState(simState *module.SimulationState) {
	var maxMemoChars uint64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, MaxMemoChars, &maxMemoChars, simState.Rand,
		func(r *rand.Rand) { maxMemoChars = GenMaxMemoChars(r) },
	)

	var txSigLimit uint64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, TxSigLimit, &txSigLimit, simState.Rand,
		func(r *rand.Rand) { txSigLimit = GenTxSigLimit(r) },
	)

	var txSizeCostPerByte uint64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, TxSizeCostPerByte, &txSizeCostPerByte, simState.Rand,
		func(r *rand.Rand) { txSizeCostPerByte = GenTxSizeCostPerByte(r) },
	)

	var sigVerifyCostED25519 uint64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, SigVerifyCostED25519, &sigVerifyCostED25519, simState.Rand,
		func(r *rand.Rand) { sigVerifyCostED25519 = GenSigVerifyCostED25519(r) },
	)

	var sigVerifyCostSECP256K1 uint64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, SigVerifyCostSECP256K1, &sigVerifyCostSECP256K1, simState.Rand,
		func(r *rand.Rand) { sigVerifyCostSECP256K1 = GenSigVerifyCostSECP256K1(r) },
	)

	params := types.NewParams(maxMemoChars, txSigLimit, txSizeCostPerByte,
		sigVerifyCostED25519, sigVerifyCostSECP256K1)
	genesisAccs := RandomGenesisAccounts(simState)

	authGenesis := types.NewGenesisState(params, genesisAccs)

	fmt.Printf("Selected randomly generated auth parameters:\n%s\n", codec.MustMarshalJSONIndent(simState.Cdc, authGenesis.Params))
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(authGenesis)
}

// RandomGenesisAccounts returns randomly generated genesis accounts
func RandomGenesisAccounts(simState *module.SimulationState) (genesisAccs exported.GenesisAccounts) {
	for i, acc := range simState.Accounts {
		coins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(simState.InitialStake))}
		bacc := types.NewBaseAccountWithAddress(acc.Address)
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
				gacc = vestingtypes.NewContinuousVestingAccount(&bacc, startTime, endTime)
			} else {
				gacc = vestingtypes.NewDelayedVestingAccount(&bacc, endTime)
			}
		}
		genesisAccs = append(genesisAccs, gacc)
	}

	return genesisAccs
}
