package stake

import (
<<<<<<< HEAD
	"fmt"
=======
>>>>>>> af8e5b35c0923b2eaebfbaf6dd22436e930e5d9f
	"math/rand"
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//change the int in NewSource to generate random input for tests that use r for randomization
var r = rand.New(rand.NewSource(505))

func TestGetInflation(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)
	pool := keeper.GetPool(ctx)
	params := keeper.GetParams(ctx)

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

//Tests that provisions are added to the pool as expected for 8766 hours (a year)
func TestProcessProvisions(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)
	params := defaultParams()
	params.MaxValidators = 2
	keeper.setParams(ctx, params)
	pool := keeper.GetPool(ctx)

	var tokenSupply int64 = 550000000
	var bondedShares int64 = 150000000
	var unbondedShares int64 = 400000000

	// create some validators some bonded, some unbonded
	var validators [5]Validator
	validators[0] = NewValidator(addrs[0], pks[0], Description{})
	validators[0], pool, _ = validators[0].addTokensFromDel(pool, 150000000)
	keeper.setPool(ctx, pool)
	validators[0] = keeper.updateValidator(ctx, validators[0])
	pool = keeper.GetPool(ctx)
	require.Equal(t, bondedShares, pool.BondedTokens, "%v", pool)

	validators[1] = NewValidator(addrs[1], pks[1], Description{})
	validators[1], pool, _ = validators[1].addTokensFromDel(pool, 100000000)
	keeper.setPool(ctx, pool)
	validators[1] = keeper.updateValidator(ctx, validators[1])
	validators[2] = NewValidator(addrs[2], pks[2], Description{})
	validators[2], pool, _ = validators[2].addTokensFromDel(pool, 100000000)
	keeper.setPool(ctx, pool)
	validators[2] = keeper.updateValidator(ctx, validators[2])
	validators[3] = NewValidator(addrs[3], pks[3], Description{})
	validators[3], pool, _ = validators[3].addTokensFromDel(pool, 100000000)
	keeper.setPool(ctx, pool)
	validators[3] = keeper.updateValidator(ctx, validators[3])
	validators[4] = NewValidator(addrs[4], pks[4], Description{})
	validators[4], pool, _ = validators[4].addTokensFromDel(pool, 100000000)
	keeper.setPool(ctx, pool)
	validators[4] = keeper.updateValidator(ctx, validators[4])

	assert.Equal(t, tokenSupply, pool.TokenSupply())
	assert.Equal(t, bondedShares, pool.BondedTokens)
	assert.Equal(t, unbondedShares, pool.UnbondedTokens)
	var provisionTallied int64 = 0

	// fmt.Printf("pool bonded ratio: %v\n", pool)
	bondedRatioTest := pool.bondedRatio()
	fmt.Println("pool bonded ratio: ", bondedRatioTest) // is 150 mil / 550 mil, or 15/55, divisor 5 is 3/11

	// 1 to 1 ratio
	var bondedShares, unbondedShares sdk.Rat = sdk.NewRat(150000000, 1), sdk.NewRat(400000000, 1)

	// initial bonded ratio ~ 27%
	assert.True(t, pool.bondedRatio().Equal(sdk.NewRat(bondedShares, tokenSupply)), "%v", pool.bondedRatio())

	bondedShareExRateTest := pool.bondedShareExRate()
	fmt.Println("pool bondedSharExTRate: ", bondedShareExRateTest)

	// test the value of candidate shares
	assert.True(t, pool.bondedShareExRate().Equal(sdk.OneRat()), "%v", pool.bondedShareExRate())

	initialSupply := pool.TokenSupply()
	initialUnbonded := pool.TokenSupply() - pool.BondedTokens

	// process the provisions a year
	for hr := 0; hr < 8766; hr++ {
		pool := keeper.GetPool(ctx)
		_, expProvisions, _ := checkAndProcessProvisions(t, keeper, pool, ctx, hr)
		cumulativeExpProvs = cumulativeExpProvs + expProvisions
	}
	pool = keeper.GetPool(ctx)
	//fmt.Printf("hr %v, startBondedPool %v, expProvisions %v, pool.BondedPool %v\n", hr, startBondedPool, expProvisions, pool.BondedPool)
	assert.NotEqual(t, initialSupply, pool.TotalSupply)
	assert.Equal(t, initialUnbonded, pool.UnbondedPool)
	//panic(fmt.Sprintf("debug total %v, bonded  %v, diff %v\n", p.TotalSupply, p.BondedPool, pool.TotalSupply-pool.BondedPool))
	fmt.Printf("debug total %v, \nbonded  %v, \ndiff %v\n", pool.TotalSupply, pool.BondedPool, pool.TotalSupply-pool.BondedPool)
	// fmt.Println("provisionTally: ", provisionTally)

	// Final check that the pool equals initial values + provisions and adjustments we recorded
	pool = keeper.GetPool(ctx)
	checkFinalPoolValues(t, pool, initialTotalTokens,
		initialUnbondedTokens, cumulativeExpProvs,
		0, 0, bondedShares, unbondedShares)
}

//Tests that the hourly rate of change will be positve, negative, or zero, depending on bonded ratio and inflation rate
//Cycles through the whole gambit of starting at 7% inflation, up to 20%, back down to 7% (it takes 11.4 years)
func TestHourlyRateOfChange(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)
	params := defaultParams()
	keeper.setParams(ctx, params)
	pool := keeper.GetPool(ctx)

	// create some candidates some bonded, some unbonded
	pool = setupCandidates(pool, keeper, ctx, 10, 0, 5)

	// test setUpCandidates returned the token values by passing these vars into checkCandidateSetup()
	var (
		initialTotalTokens    int64 = 550000000
		initialBondedTokens   int64 = 150000000
		initialUnbondedTokens int64 = 400000000
		cumulativeExpProvs    int64
		bondedShares          = sdk.NewRat(150000000, 1)
		unbondedShares        = sdk.NewRat(400000000, 1)
	)
	checkCandidateSetup(t, pool, initialTotalTokens, initialBondedTokens, initialUnbondedTokens)

	// process the provisions a year
	for hr := 0; hr < 8766; hr++ {
		pool := keeper.GetPool(ctx)
		_, expProvisions, _ := checkAndProcessProvisions(t, keeper, pool, ctx, hr)
		cumulativeExpProvs = cumulativeExpProvs + expProvisions
	}

	// Final check that the pool equals initial values + provisions and adjustments we recorded
	pool = keeper.GetPool(ctx)
	checkFinalPoolValues(t, pool, initialTotalTokens,
		initialUnbondedTokens, cumulativeExpProvs,
		0, 0, bondedShares, unbondedShares)
}

//Test that a large unbonding will significantly lower the bonded ratio
func TestLargeUnbond(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)
	params := defaultParams()
	keeper.setParams(ctx, params)
	pool := keeper.GetPool(ctx)

	// Candidates unbonded (0-4), bonded (5-9),
	// candidate 9 will be unbonded, bringing us from ~73% to ~55%
	pool = setupCandidates(pool, keeper, ctx, 10, 5, 10)

	// test setUpCandidates returned the token values by passing these vars into checkCandidateSetup()
	var (
		initialTotalTokens    int64 = 550000000
		initialBondedTokens   int64 = 400000000
		initialUnbondedTokens int64 = 150000000
		cand9UnbondedTokens   int64
		bondedShares          = sdk.NewRat(400000000, 1)
		unbondedShares        = sdk.NewRat(150000000, 1)
		bondSharesCand9       = sdk.NewRat(100000000, 1)
	)
	checkCandidateSetup(t, pool, initialTotalTokens, initialBondedTokens, initialUnbondedTokens)

	pool = keeper.GetPool(ctx)
	candidate, found := keeper.GetCandidate(ctx, addrs[9])
	assert.True(t, found)

	// initialBondedRatio that we can use to compare to the new values after the unbond
	initialBondedRatio := pool.bondedRatio()

	// This func will unbond 100,000,000 tokens that were previously bonded
	pool, candidate, _, _ = OpBondOrUnbond(r, pool, candidate)
	keeper.setPool(ctx, pool)

	// process provisions after the bonding, to compare the difference in expProvisions and expInflation
	_, expProvisionsAfter, pool := checkAndProcessProvisions(t, keeper, pool, ctx, 0)

	bondedShares = bondedShares.Sub(bondSharesCand9)
	cand9UnbondedTokens = pool.unbondedShareExRate().Mul(candidate.Assets).Evaluate()
	unbondedShares = unbondedShares.Add(sdk.NewRat(cand9UnbondedTokens, 1).Mul(pool.unbondedShareExRate()))

	// unbonded shares should increase
	assert.True(t, unbondedShares.GT(sdk.NewRat(150000000, 1)))
	// Ensure that new bonded ratio is less than old bonded ratio , because before they were increasing (i.e. 55 < 72)
	assert.True(t, (pool.bondedRatio().LT(initialBondedRatio)))

	// Final check that the pool equals initial values + provisions and adjustments we recorded
	pool = keeper.GetPool(ctx)
	checkFinalPoolValues(t, pool, initialTotalTokens,
		initialUnbondedTokens, expProvisionsAfter,
		-cand9UnbondedTokens, cand9UnbondedTokens, bondedShares, unbondedShares)
}

//Test that a large bonding will cause inflation to go down, and lower the bondedRatio
func TestLargeBond(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)
	params := defaultParams()
	keeper.setParams(ctx, params)
	pool := keeper.GetPool(ctx)

	// Candidates unbonded (0-4), bonded (5-8), candidate 9 left unbonded, so it can be bonded later in the test
	// bonded candidate 9 brings us from ~55% to ~73 bondedRatio
	pool = setupCandidates(pool, keeper, ctx, 10, 5, 9)

	// test setUpCandidates returned the token values by passing these vars into checkCandidateSetup()
	var (
		initialTotalTokens    int64 = 550000000
		initialBondedTokens   int64 = 300000000
		initialUnbondedTokens int64 = 250000000
		cand9unbondedTokens   int64 = 100000000
		cand9bondedTokens     int64
		bondedShares          = sdk.NewRat(300000000, 1)
		unbondedShares        = sdk.NewRat(250000000, 1)
		unbondSharesCand9     = sdk.NewRat(100000000, 1)
	)
	checkCandidateSetup(t, pool, initialTotalTokens, initialBondedTokens, initialUnbondedTokens)

	pool = keeper.GetPool(ctx)
	candidate, found := keeper.GetCandidate(ctx, addrs[9])
	assert.True(t, found)

	// initialBondedRatio that we can use to compare to the new values after the unbond
	initialBondedRatio := pool.bondedRatio()

	// This func will bond 100,000,000 tokens that were previously unbonded
	pool, candidate, _, _ = OpBondOrUnbond(r, pool, candidate)
	keeper.setPool(ctx, pool)

	// process provisions after the bonding, to compare the difference in expProvisions and expInflation
	_, expProvisionsAfter, pool := checkAndProcessProvisions(t, keeper, pool, ctx, 0)

	unbondedShares = unbondedShares.Sub(unbondSharesCand9)
	cand9bondedTokens = cand9unbondedTokens
	cand9unbondedTokens = 0
	bondedTokens := initialBondedTokens + cand9bondedTokens + expProvisionsAfter
	bondedShares = sdk.NewRat(bondedTokens, 1).Quo(pool.bondedShareExRate())

	// bonded shares should increase
	assert.True(t, bondedShares.GT(sdk.NewRat(300000000, 1)))
	//Ensure that new bonded ratio is greater than old bonded ratio, since we just added 100,000 bonded
	assert.True(t, (pool.bondedRatio().GT(initialBondedRatio)))

	// Final check that the pool equals initial values + provisions and adjustments we recorded
	checkFinalPoolValues(t, pool, initialTotalTokens,
		initialUnbondedTokens, expProvisionsAfter,
		cand9bondedTokens, -cand9bondedTokens, bondedShares, unbondedShares)
}

//Tests that the hourly rate of change will be positve, negative, or zero, depending on bonded ratio and inflation rate
//Cycles through the whole gambit of starting at 7% inflation, up to 20%, back down to 7% (it takes ~11.4 years)
func TestHourlyInflationRateOfChange(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)
	params := defaultParams()
	keeper.setParams(ctx, params)
	pool := keeper.GetPool(ctx)

	// create some candidates some bonded, some unbonded
	pool = setupCandidates(pool, keeper, ctx, 10, 0, 5)

	// test setUpCandidates returned the token values by passing these vars into checkCandidateSetup()
	var (
		initialTotalTokens    int64 = 550000000
		initialBondedTokens   int64 = 150000000
		initialUnbondedTokens int64 = 400000000
		cumulativeExpProvs    int64
		bondedShares          = sdk.NewRat(150000000, 1)
		unbondedShares        = sdk.NewRat(400000000, 1)
	)
	checkCandidateSetup(t, pool, initialTotalTokens, initialBondedTokens, initialUnbondedTokens)

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

//Tests that inflation works as expected when we get do a random operation on 20 different candidates
func TestInflationWithRandomOperations(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)
	params := defaultParams()
	keeper.setParams(ctx, params)
	numCandidates := 20 //max 40 possible right now since var addrs only goes up to addrs[39]

	//start off by randomly creating 20 candidates
	pool, candidates := randomSetup(r, numCandidates)
	require.Equal(t, numCandidates, len(candidates))

	for i := 0; i < len(candidates); i++ {
		keeper.setCandidate(ctx, candidates[i])
	}
	keeper.setPool(ctx, pool)

	//This counter is used to rotate through each candidate so each random operation is applied to a different candidate
	candidateCounter := 0

	// looping through 20 random operations
	for i := 0; i < numCandidates; i++ {
		pool := keeper.GetPool(ctx)

		//Get inflation before randomOperation
		previousInflation := pool.Inflation

		//Random operation, and recording how candidates are modified
		poolMod, candidateMod, tokens, msg := randomOperation(r)(r, pool, candidates[candidateCounter])
		candidatesMod := make([]Candidate, len(candidates))
		copy(candidatesMod[:], candidates[:])
		require.Equal(t, numCandidates, len(candidates), "i %v", candidateCounter)
		require.Equal(t, numCandidates, len(candidatesMod), "i %v", candidateCounter)
		candidatesMod[candidateCounter] = candidateMod

		assertInvariants(t, msg,
			pool, candidates,
			poolMod, candidatesMod, tokens)

		//set pool and candidates after the random operation
		pool = poolMod
		keeper.setPool(ctx, pool)
		candidates = candidatesMod

		//Must set inflation here, as opposed to when we have a test where we call processProvisions(), which updates pool.Inflation
		updatedInflation := keeper.nextInflation(ctx)
		pool.Inflation = updatedInflation
		keeper.setPool(ctx, pool)

		checkInflation(t, pool, previousInflation, updatedInflation, msg)
		candidateCounter++
	}
}

////////////////////////////////HELPER FUNCTIONS BELOW/////////////////////////////////////

// Final check on the global pool values against what each test added up hour by hour.
// bondedAdjustment and unbondedAdjustment are the calculated changes
// that the test calling this function accumlated (i.e. if three unbonds happened, their total value passed as unbondedAdjustment)
func checkFinalPoolValues(t *testing.T, pool Pool, initialTotalTokens, initialUnbondedTokens,
	cumulativeExpProvs, bondedAdjustment, unbondedAdjustment int64, bondedShares, unbondedShares sdk.Rat) {

	initialBonded := initialTotalTokens - initialUnbondedTokens
	calculatedTotalTokens := initialTotalTokens + cumulativeExpProvs
	calculatedBondedTokens := initialBonded + cumulativeExpProvs + bondedAdjustment
	calculatedUnbondedTokens := initialUnbondedTokens + unbondedAdjustment

	// test that the bonded ratio the pool has is equal to what we calculated for tokens
	assert.True(t, pool.bondedRatio().Equal(sdk.NewRat(calculatedBondedTokens, calculatedTotalTokens)), "%v", pool.bondedRatio())

	// test global supply
	assert.Equal(t, calculatedTotalTokens, pool.TotalSupply)
	assert.Equal(t, calculatedBondedTokens, pool.BondedPool)
	assert.Equal(t, calculatedUnbondedTokens, pool.UnbondedPool)

	// test the value of candidate shares
	assert.True(t, pool.bondedShareExRate().Mul(bondedShares).Equal(sdk.NewRat(calculatedBondedTokens)), "%v", pool.bondedShareExRate())
	assert.True(t, pool.unbondedShareExRate().Mul(unbondedShares).Equal(sdk.NewRat(calculatedUnbondedTokens)), "%v", pool.unbondedShareExRate())
}

// Checks provisions are added to the pool correctly every hour
// Returns expected Provisions, expected Inflation, and pool, to help with cumulative calculations back in main Tests
func checkAndProcessProvisions(t *testing.T, keeper Keeper, pool Pool, ctx sdk.Context, hr int) (sdk.Rat, int64, Pool) {

	//If we are not doing a random operation, just check that normal provisions are working for each hour
	expInflation := keeper.nextInflation(ctx)
	expProvisions := (expInflation.Mul(sdk.NewRat(pool.TotalSupply)).Quo(hrsPerYrRat)).Evaluate()
	// expPRoTemp := (expInflation.Mul(sdk.NewRat(pool.TotalSupply)).Quo(hrsPerYrRat))
	startBondedPool := pool.BondedPool
	startTotalSupply := pool.TotalSupply
	pool = keeper.processProvisions(ctx)
	keeper.setPool(ctx, pool)
	// fmt.Println("expprotemp: ", expPRoTemp)

	//check provisions were added to pool
	require.Equal(t, startBondedPool+expProvisions, pool.BondedPool, "hr %v", hr)
	require.Equal(t, startTotalSupply+expProvisions, pool.TotalSupply)

	return expInflation, expProvisions, pool
}

// Deterministic setup of candidates
// Allows you to decide how many candidates to setup, and which ones you want bonded
// Tokens allocated to each candidate increase by 10000000 for each candidate
func setupCandidates(pool Pool, keeper Keeper, ctx sdk.Context, numCands, indexBondedGT, indexBondedLT int) Pool {

	candidates := make([]Candidate, numCands)
	for i := 0; i < numCands; i++ {
		c := Candidate{
			Status:      Unbonded,
			PubKey:      pks[i],
			Address:     addrs[i],
			Assets:      sdk.NewRat(0),
			Liabilities: sdk.NewRat(0),
		}
		if i >= indexBondedGT && i < indexBondedLT {
			c.Status = Bonded
		}
		mintedTokens := int64((i + 1) * 10000000)
		pool.TotalSupply += mintedTokens
		pool, c, _ = pool.candidateAddTokens(c, mintedTokens)

		keeper.setCandidate(ctx, c)
		candidates[i] = c
	}
	keeper.setPool(ctx, pool)
	pool = keeper.GetPool(ctx)
	return pool
}

// Checks that the deterministic candidate setup you wanted matches the values in the pool
func checkCandidateSetup(t *testing.T, pool Pool, initialTotalTokens, initialBondedTokens, initialUnbondedTokens int64) {

	assert.Equal(t, initialTotalTokens, pool.TotalSupply)
	assert.Equal(t, initialBondedTokens, pool.BondedPool)
	assert.Equal(t, initialUnbondedTokens, pool.UnbondedPool)

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

//TODO: fix update inflation and expNext or whatever, seem to be usingit twice
//TODO: make all of the variables named the same
