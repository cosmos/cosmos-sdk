package simulation

import (
	"encoding/json"
	"math/rand"
	"time"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/x/auth/types"
	vestingtypes "cosmossdk.io/x/auth/vesting/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
)

// RandomGenesisAccounts defines the default RandomGenesisAccountsFn used on the SDK.
// It creates a slice of BaseAccount, ContinuousVestingAccount and DelayedVestingAccount.
func RandomGenesisAccounts(r *rand.Rand, bondDenom string, accs []simulation.Account, initialStake sdkmath.Int, numBonded int, genTimestamp time.Time) types.GenesisAccounts {
	genesisAccs := make(types.GenesisAccounts, len(accs))
	for i, acc := range accs {
		bacc := types.NewBaseAccountWithAddress(acc.Address)

		// Only consider making a vesting account once the initial bonded validator
		// set is exhausted due to needing to track DelegatedVesting.
		if !(i > numBonded && r.Intn(100) < 50) {
			genesisAccs[i] = bacc
			continue
		}

		initialVesting := sdk.NewCoins(sdk.NewInt64Coin(bondDenom, r.Int63n(initialStake.Int64())))
		var endTime int64

		startTime := genTimestamp.Unix()

		// Allow for some vesting accounts to vest very quickly while others very slowly.
		if r.Intn(100) < 50 {
			endTime = int64(simulation.RandIntBetween(r, int(startTime)+1, int(startTime+(60*60*24*30))))
		} else {
			endTime = int64(simulation.RandIntBetween(r, int(startTime)+1, int(startTime+(60*60*12))))
		}

		bva, err := vestingtypes.NewBaseVestingAccount(bacc, initialVesting, endTime)
		if err != nil {
			panic(err)
		}

		if r.Intn(100) < 50 {
			genesisAccs[i] = vestingtypes.NewContinuousVestingAccountRaw(bva, startTime)
		} else {
			genesisAccs[i] = vestingtypes.NewDelayedVestingAccountRaw(bva)
		}
	}

	return genesisAccs
}

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
func RandomizedGenState(r *rand.Rand, genState map[string]json.RawMessage, cdc codec.JSONCodec, bondDenom string, accs []simulation.Account, initialStake sdkmath.Int, numBonded int, genTimestamp time.Time) {
	maxMemoChars := GenMaxMemoChars(r)
	txSigLimit := GenTxSigLimit(r)
	txSizeCostPerByte := GenTxSizeCostPerByte(r)
	sigVerifyCostED25519 := GenSigVerifyCostED25519(r)
	sigVerifyCostSECP256K1 := GenSigVerifyCostSECP256K1(r)

	params := types.NewParams(maxMemoChars, txSigLimit, txSizeCostPerByte,
		sigVerifyCostED25519, sigVerifyCostSECP256K1)
	genesisAccs := RandomGenesisAccounts(r, bondDenom, accs, initialStake, numBonded, genTimestamp)

	authGenesis := types.NewGenesisState(params, genesisAccs)

	genState[types.ModuleName] = cdc.MustMarshalJSON(authGenesis)
}
