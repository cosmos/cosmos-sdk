package stake

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetInflation(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)
	pool := keeper.GetPool(ctx)
	params := keeper.GetParams(ctx)
	hrsPerYrRat := sdk.NewRat(hrsPerYr)

	// Governing Mechanism:
	//    bondedRatio = BondedTokens / TotalSupply
	//    inflationRateChangePerYear = (1- bondedRatio/ GoalBonded) * MaxInflationRateChange

	tests := []struct {
		name                            string
		setBondedTokens, setLooseTokens int64
		setInflation, expectedChange    sdk.Rat
	}{
		// with 0% bonded atom supply the inflation should increase by InflationRateChange
		{"test 1", 0, 0, sdk.NewRat(7, 100), params.InflationRateChange.Quo(hrsPerYrRat).Round(precision)},

		// 100% bonded, starting at 20% inflation and being reduced
		// (1 - (1/0.67))*(0.13/8667)
		{"test 2", 1, 0, sdk.NewRat(20, 100),
			sdk.OneRat().Sub(sdk.OneRat().Quo(params.GoalBonded)).Mul(params.InflationRateChange).Quo(hrsPerYrRat).Round(precision)},

		// 50% bonded, starting at 10% inflation and being increased
		{"test 3", 1, 1, sdk.NewRat(10, 100),
			sdk.OneRat().Sub(sdk.NewRat(1, 2).Quo(params.GoalBonded)).Mul(params.InflationRateChange).Quo(hrsPerYrRat).Round(precision)},

		// test 7% minimum stop (testing with 100% bonded)
		{"test 4", 1, 0, sdk.NewRat(7, 100), sdk.ZeroRat()},
		{"test 5", 1, 0, sdk.NewRat(70001, 1000000), sdk.NewRat(-1, 1000000).Round(precision)},

		// test 20% maximum stop (testing with 0% bonded)
		{"test 6", 0, 0, sdk.NewRat(20, 100), sdk.ZeroRat()},
		{"test 7", 0, 0, sdk.NewRat(199999, 1000000), sdk.NewRat(1, 1000000).Round(precision)},

		// perfect balance shouldn't change inflation
		{"test 8", 67, 33, sdk.NewRat(15, 100), sdk.ZeroRat()},
	}
	for _, tc := range tests {
		pool.BondedTokens, pool.LooseUnbondedTokens = tc.setBondedTokens, tc.setLooseTokens
		pool.Inflation = tc.setInflation
		keeper.setPool(ctx, pool)

		inflation := keeper.nextInflation(ctx)
		diffInflation := inflation.Sub(tc.setInflation)

		assert.True(t, diffInflation.Equal(tc.expectedChange),
			"Name: %v\nDiff:  %v\nExpected: %v\n", tc.name, diffInflation, tc.expectedChange)
	}
}

func TestProcessProvisions(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)
	pool := keeper.GetPool(ctx)

	var (
		initialTotalTokens    int64 = 550000000
		initialBondedTokens   int64 = 250000000
		initialUnbondedTokens int64 = 300000000
		cumulativeExpProvs    int64
		initialBondedShares          = sdk.NewRat(250000000, 1)
		initialUnbondedShares        = sdk.NewRat(300000000, 1)
		tokensForValidators          = []int64{150000000, 100000000, 100000000, 100000000, 100000000}
		bondedValidators      uint16 = 2
	)

	// create some validators some bonded, some unbonded
	_, keeper, pool = setupTestValidators(pool, keeper, ctx, tokensForValidators, bondedValidators)
	checkValidatorSetup(t, pool, initialTotalTokens, initialBondedTokens, initialUnbondedTokens)

	// process the provisions for a year
	for hr := 0; hr < 8766; hr++ {
		pool := keeper.GetPool(ctx)
		_, expProvisions, _ := checkAndProcessProvisions(t, keeper, pool, ctx, hr)
		cumulativeExpProvs = cumulativeExpProvs + expProvisions
	}

	//get the pool and do the final value checks from checkFinalPoolValues
	pool = keeper.GetPool(ctx)
	checkFinalPoolValues(t, pool, initialTotalTokens,
		initialUnbondedTokens, cumulativeExpProvs,
		0, 0, initialBondedShares, initialUnbondedShares)
}

// Tests that the hourly rate of change will be positive, negative, or zero, depending on bonded ratio and inflation rate
// Cycles through the whole gambit of starting at 7% inflation, up to 20%, back down to 7% (it takes ~11.4 years)
func TestHourlyInflationRateOfChange(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)
	// params := defaultParams()
	// keeper.setParams(ctx, params)
	pool := keeper.GetPool(ctx)

	// test setUpCandidates returned the token values by passing these vars into checkCandidateSetup()
	var (
		initialTotalTokens    int64 = 550000000
		initialBondedTokens   int64 = 150000000
		initialUnbondedTokens int64 = 400000000
		cumulativeExpProvs    int64
		bondedShares                 = sdk.NewRat(150000000, 1)
		unbondedShares               = sdk.NewRat(400000000, 1)
		tokensForValidators          = []int64{150000000, 100000000, 100000000, 100000000, 100000000}
		bondedValidators      uint16 = 1
	)

	// create some candidates some bonded, some unbonded
	_, keeper, pool = setupTestValidators(pool, keeper, ctx, tokensForValidators, bondedValidators)
	checkValidatorSetup(t, pool, initialTotalTokens, initialBondedTokens, initialUnbondedTokens)

	// ~11.4 years to go from 7%, up to 20%, back down to 7%
	for hr := 0; hr < 100000; hr++ {
		pool := keeper.GetPool(ctx)
		previousInflation := pool.Inflation
		updatedInflation, expProvisions, pool := checkAndProcessProvisions(t, keeper, pool, ctx, hr)
		cumulativeExpProvs = cumulativeExpProvs + expProvisions
		msg := strconv.Itoa(hr)
		checkInflation(t, pool, previousInflation, updatedInflation, msg)
	}

	// Final check that the pool equals initial values + cumulative provisions and adjustments we recorded
	pool = keeper.GetPool(ctx)
	checkFinalPoolValues(t, pool, initialTotalTokens,
		initialUnbondedTokens, cumulativeExpProvs,
		0, 0, bondedShares, unbondedShares)
}

////////////////////////////////HELPER FUNCTIONS BELOW/////////////////////////////////////

// Final check on the global pool values for what the total tokens accumulated from each hour of provisions and other functions
// bondedAdjustment and unbondedAdjustment are the accumulated changes for the operations of the test (i.e. if three unbonds happened, their total value would be passed as unbondedAdjustment)
func checkFinalPoolValues(t *testing.T, pool Pool, initialTotalTokens, initialUnbondedTokens,
	cumulativeExpProvs, bondedAdjustment, unbondedAdjustment int64, bondedShares, unbondedShares sdk.Rat) {

	initialBonded := initialTotalTokens - initialUnbondedTokens
	calculatedTotalTokens := initialTotalTokens + cumulativeExpProvs
	calculatedBondedTokens := initialBonded + cumulativeExpProvs + bondedAdjustment
	calculatedUnbondedTokens := initialUnbondedTokens + unbondedAdjustment

	// test that the bonded ratio the pool has is equal to what we calculated for tokens
	assert.True(t, pool.bondedRatio().Equal(sdk.NewRat(calculatedBondedTokens, calculatedTotalTokens)), "%v", pool.bondedRatio())

	// test global supply
	assert.Equal(t, calculatedTotalTokens, pool.TokenSupply())
	assert.Equal(t, calculatedBondedTokens, pool.BondedTokens)
	assert.Equal(t, calculatedUnbondedTokens, pool.UnbondedTokens)

	// test the value of candidate shares
	assert.True(t, pool.bondedShareExRate().Mul(bondedShares).Equal(sdk.NewRat(calculatedBondedTokens)), "%v", pool.bondedShareExRate())
	assert.True(t, pool.unbondedShareExRate().Mul(unbondedShares).Equal(sdk.NewRat(calculatedUnbondedTokens)), "%v", pool.unbondedShareExRate())
}

// Checks provisions are added to the pool correctly every hour
// Returns expected Provisions, expected Inflation, and pool, to help with cumulative calculations back in main Tests
func checkAndProcessProvisions(t *testing.T, keeper Keeper, pool Pool, ctx sdk.Context, hr int) (sdk.Rat, int64, Pool) {

	//If we are not doing a random operation, just check that normal provisions are working for each hour
	expInflation := keeper.nextInflation(ctx)
	expProvisions := (expInflation.Mul(sdk.NewRat(pool.TokenSupply())).Quo(hrsPerYrRat)).Evaluate()
	startBondedPool := pool.BondedTokens
	startTotalSupply := pool.TokenSupply()
	pool = keeper.processProvisions(ctx)
	keeper.setPool(ctx, pool)

	//check provisions were added to pool
	require.Equal(t, startBondedPool+expProvisions, pool.BondedTokens, "hr %v", hr)
	require.Equal(t, startTotalSupply+expProvisions, pool.TokenSupply())

	return expInflation, expProvisions, pool
}

// Deterministic setup of validators, which updates the pool and choose maxValidators to be bonded
// Allows you to decide how many validators to setup, and which ones you want bonded
// You choose bonded validators by setting params.MaxValidators. If you choose 2, the first 2 Validators in the arrray will be bonded, the rest unbonded
func setupTestValidators(pool Pool, keeper Keeper, ctx sdk.Context, validatorTokens []int64, maxValidators uint16) ([]Validator, Keeper, Pool) {
	params := defaultParams()
	params.MaxValidators = maxValidators //set to limit the amount of validators we want bonded
	keeper.setParams(ctx, params)
	numValidators := len(validatorTokens)
	validators := make([]Validator, numValidators)

	for i := 0; i < numValidators; i++ {
		validators[i] = NewValidator(addrs[i], pks[i], Description{})
		validators[i], pool, _ = validators[i].addTokensFromDel(pool, validatorTokens[i])
		keeper.setPool(ctx, pool)
		validators[i] = keeper.updateValidator(ctx, validators[i])
		pool = keeper.GetPool(ctx)
	}

	return validators, keeper, pool
}

// Checks that the deterministic candidate setup you wanted matches the values in the pool
func checkValidatorSetup(t *testing.T, pool Pool, initialTotalTokens, initialBondedTokens, initialUnbondedTokens int64) {

	assert.Equal(t, initialTotalTokens, pool.TokenSupply())
	assert.Equal(t, initialBondedTokens, pool.BondedTokens)
	assert.Equal(t, initialUnbondedTokens, pool.UnbondedTokens)

	// test initial bonded ratio
	assert.True(t, pool.bondedRatio().Equal(sdk.NewRat(initialBondedTokens, initialTotalTokens)), "%v", pool.bondedRatio())
	// test the value of candidate shares
	assert.True(t, pool.bondedShareExRate().Equal(sdk.OneRat()), "%v", pool.bondedShareExRate())
}

// Checks that The inflation will correctly increase or decrease after an update to the pool
func checkInflation(t *testing.T, pool Pool, previousInflation, updatedInflation sdk.Rat, msg string) {

	inflationChange := updatedInflation.Sub(previousInflation)

	switch {
	//BELOW 67% - Rate of change positive and increasing, while we are between 7% <= and < 20% inflation
	case pool.bondedRatio().LT(sdk.NewRat(67, 100)) && updatedInflation.LT(sdk.NewRat(20, 100)):
		assert.Equal(t, true, inflationChange.GT(sdk.ZeroRat()), msg)

	//BELOW 67% - Rate of change should be 0 while inflation continually stays at 20% until we reach 67% bonded ratio
	case pool.bondedRatio().LT(sdk.NewRat(67, 100)) && updatedInflation.Equal(sdk.NewRat(20, 100)):
		if previousInflation.Equal(sdk.NewRat(20, 100)) {
			assert.Equal(t, true, inflationChange.IsZero(), msg)
			//This else statement covers the one off case where we first hit 20%, but we still needed a positive ROC to get to 67% bonded ratio (i.e. we went from 19.99999% to 20%)
		} else {
			assert.Equal(t, true, inflationChange.GT(sdk.ZeroRat()), msg)
		}

	//ABOVE 67% - Rate of change should be negative while the bond is above 67, and should stay negative until we reach inflation of 7%
	case pool.bondedRatio().GT(sdk.NewRat(67, 100)) && updatedInflation.LT(sdk.NewRat(20, 100)) && updatedInflation.GT(sdk.NewRat(7, 100)):
		assert.Equal(t, true, inflationChange.LT(sdk.ZeroRat()), msg)

	//ABOVE 67% - Rate of change should be 0 while inflation continually stays at 7%.
	case pool.bondedRatio().GT(sdk.NewRat(67, 100)) && updatedInflation.Equal(sdk.NewRat(7, 100)):
		if previousInflation.Equal(sdk.NewRat(7, 100)) {
			assert.Equal(t, true, inflationChange.IsZero(), msg)
			//This else statement covers the one off case where we first hit 7%, but we still needed a negative ROC to continue to get down to 67%. (i.e. we went from 7.00001% to 7%)
		} else {
			assert.Equal(t, true, inflationChange.LT(sdk.ZeroRat()), msg)
		}
	}

}
