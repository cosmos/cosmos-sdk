package stake

import (
	"fmt"
	"math/rand"
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
		_, expProvisions, _ := checkHourlyProvisions(t, keeper, pool, ctx, hr)
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
func TestHourlyRateOfChange(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)
	params := defaultParams()
	keeper.setParams(ctx, params)
	pool := keeper.GetPool(ctx)

	// create some candidates some bonded, some unbonded
	candidates := make([]Candidate, 10)
	for i := 0; i < 10; i++ {
		c := Candidate{
			Status:      Unbonded,
			PubKey:      pks[i],
			Address:     addrs[i],
			Assets:      sdk.NewRat(0),
			Liabilities: sdk.NewRat(0),
		}
		if i < 5 {
			c.Status = Bonded
		}
		mintedTokens := int64((i + 1) * 10000000)
		pool.TotalSupply += mintedTokens
		pool, c, _ = pool.candidateAddTokens(c, mintedTokens)

		keeper.setCandidate(ctx, c)
		candidates[i] = c
	}
	keeper.setPool(ctx, pool)
	var totalSupply int64 = 550000000
	var bondedTokens int64 = 150000000
	var unbondedTokens int64 = 400000000
	var cumulativeExpProvs int64
	assert.Equal(t, totalSupply, pool.TotalSupply)
	assert.Equal(t, bondedTokens, pool.BondedPool)
	assert.Equal(t, unbondedTokens, pool.UnbondedPool)

	// 1 to 1 ratio
	var bondedShares, unbondedShares sdk.Rat = sdk.NewRat(150000000, 1), sdk.NewRat(400000000, 1)

	// initial bonded ratio ~ 27%
	assert.True(t, pool.bondedRatio().Equal(sdk.NewRat(bondedTokens, totalSupply)), "%v", pool.bondedRatio())
	// test the value of candidate shares
	assert.True(t, pool.bondedShareExRate().Equal(sdk.OneRat()), "%v", pool.bondedShareExRate())

	initialTotalTokens := pool.TotalSupply
	initialUnbondedTokens := pool.TotalSupply - pool.BondedPool

	// ~11.4 years to go from 7%, up to 20%, back down to 7%
	for hr := 0; hr < 100000; hr++ {
		pool := keeper.GetPool(ctx)

		previousInflation := pool.Inflation
		expInflation, expProvisions, pool := checkHourlyProvisions(t, keeper, pool, ctx, hr)
		cumulativeExpProvs = cumulativeExpProvs + expProvisions

		updatedInflation := pool.Inflation
		inflationChange := updatedInflation.Sub(previousInflation)
		// fmt.Println("Inflation change: ", inflationChange)
		// pbr2 := pool.bondedRatio()

		//Rate of change positive and increasing, while we are between 7% and 20% inflation
		if pool.bondedRatio().LT(sdk.NewRat(67, 100)) && expInflation.LT(sdk.NewRat(20, 100)) {
			// fmt.Println("ROCROC 1: ", inflationChange)
			assert.Equal(t, true, inflationChange.GT(sdk.ZeroRat()), strconv.Itoa(hr))
		}

		//Rate of change should be 0 while it holds at 20% a year, until we reach 67%
		if pool.bondedRatio().LT(sdk.NewRat(67, 100)) && expInflation.Equal(sdk.NewRat(20, 100)) {
			// fmt.Println("ROCROC 2: ", inflationChange)

			if previousInflation.Equal(sdk.NewRat(20, 100)) {
				assert.Equal(t, true, inflationChange.IsZero(), strconv.Itoa(hr))
				//This covers the one off case where we first hit 20%, but we still needed a positive ROC to get there
			} else {
				assert.Equal(t, true, inflationChange.GT(sdk.ZeroRat()), strconv.Itoa(hr))
			}
		}

		//Rate of change should be negative while the bond is above 67, and should stay negative until we reach inflation of 7%
		if pool.bondedRatio().GT(sdk.NewRat(67, 100)) && expInflation.LT(sdk.NewRat(20, 100)) && expInflation.GT(sdk.NewRat(7, 100)) {
			// fmt.Println("ROCROC 3: ", inflationChange)
			assert.Equal(t, true, inflationChange.LT(sdk.ZeroRat()), strconv.Itoa(hr))
		}

		//Rate of change should be 0 while we hold at 7%.
		if pool.bondedRatio().GT(sdk.NewRat(67, 100)) && expInflation.Equal(sdk.NewRat(7, 100)) {

			if previousInflation.Equal(sdk.NewRat(7, 100)) {
				assert.Equal(t, true, inflationChange.IsZero(), strconv.Itoa(hr))
				//This covers the one off case where we first hit 7%, but we still needed a negative ROC to get there
			} else {
				assert.Equal(t, true, inflationChange.LT(sdk.ZeroRat()), strconv.Itoa(hr))

			}
			// fmt.Println("ROCROC 4: ", inflationChange)

		}
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

	// create some candidates some bonded, some unbonded. the largest candidates are bonded
	// so that we can have a big change in bonded ratio (~73% to ~55%) when cand9 unbonds
	candidates := make([]Candidate, 10)
	for i := 0; i < 10; i++ {
		c := Candidate{
			Status:      Unbonded,
			PubKey:      pks[i],
			Address:     addrs[i],
			Assets:      sdk.NewRat(0),
			Liabilities: sdk.NewRat(0),
		}
		if i > 4 {
			c.Status = Bonded
		}
		mintedTokens := int64((i + 1) * 10000000)
		pool.TotalSupply += mintedTokens
		pool, c, _ = pool.candidateAddTokens(c, mintedTokens)
		keeper.setCandidate(ctx, c)
		candidates[i] = c
	}
	keeper.setPool(ctx, pool)

	//initialzing variables for candidate state. These will get updated throughout the test
	var totalSupply, bondedTokens, unbondedTokens, cand9UnbondedTokens int64 = 550000000, 400000000, 150000000, 0
	//share to token ratio starts off as 1:1
	var bondedShares, unbondedShares, bondSharesCand9 sdk.Rat = sdk.NewRat(400000000, 1), sdk.NewRat(150000000, 1), sdk.NewRat(100000000, 1)
	// cumulative count of provisions, so we can check at the end of the year that the bonded pool has increased by this amount
	var cumulativeExpProvs int64

	assert.Equal(t, totalSupply, pool.TotalSupply)
	assert.Equal(t, bondedTokens, pool.BondedPool)
	assert.Equal(t, unbondedTokens, pool.UnbondedPool)

	// initial bonded ratio ~ 72%
	assert.True(t, pool.bondedRatio().Equal(sdk.NewRat(bondedTokens, totalSupply)), "%v", pool.bondedRatio())

	// test the value of candidate shares. should start off 1:1
	assert.True(t, pool.bondedShareExRate().Equal(sdk.OneRat()), "%v", pool.bondedShareExRate())

	initialTotalTokens := pool.TotalSupply
	initialUnbondedTokens := pool.TotalSupply - pool.BondedPool

	// process the provisions a year
	for hr := 0; hr < 8766; hr++ {
		pool := keeper.GetPool(ctx)
		_, expProvisions, pool := checkHourlyProvisions(t, keeper, pool, ctx, hr)
		cumulativeExpProvs = cumulativeExpProvs + expProvisions

		//hour 1600 was arbitrarily picked to unbond the largest candidate, and onwards of 1600 the pool.UnbondedPool amount will be larger
		if hr <= 1600 {
			require.Equal(t, unbondedTokens, pool.UnbondedPool)
		} else {
			require.Equal(t, initialUnbondedTokens+cand9UnbondedTokens, pool.UnbondedPool)
		}

		//inside this if statement are the steps to create unbonding of a candidate at hour 1600, to drop bonded ratio from ~72% to ~55%
		if hr == 1600 {
			candidate, found := keeper.GetCandidate(ctx, addrs[9])
			assert.True(t, found)
			beforeBondedRatio := pool.bondedRatio()

			//unbond 100,000,000 tokens, plus what was accumulated from provisions over 1600 hours, roughly 1,700,000
			pool, candidate = pool.bondedToUnbondedPool(candidate)
			keeper.setPool(ctx, pool)

			//bonded shares stay the same, bonded tokens have increased, meaning candidate 9 will have a favorable token to share ratio
			bondedShares = bondedShares.Sub(bondSharesCand9)
			cand9UnbondedTokens = pool.unbondedShareExRate().Mul(candidate.Assets).Evaluate()
			bondedTokens = bondedTokens - cand9UnbondedTokens

			//unbonded shares will increase
			unbondedShares = unbondedShares.Add(sdk.NewRat(cand9UnbondedTokens, 1).Mul(pool.unbondedShareExRate()))

			//Ensure that new bonded ratio is less than old bonded ratio , because before they were increasing (i.e. 55 < 72)
			assert.True(t, (pool.bondedRatio().LT(beforeBondedRatio)))
		}
	}

	// Final check that the pool equals initial values + provisions and adjustments we recorded
	pool = keeper.GetPool(ctx)
	checkFinalPoolValues(t, pool, initialTotalTokens,
		initialUnbondedTokens, cumulativeExpProvs,
		-cand9UnbondedTokens, cand9UnbondedTokens, bondedShares, unbondedShares)

}

//Test that a large bonding will significantly lower the bonded ratio
func TestLargeBond(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)
	params := defaultParams()
	keeper.setParams(ctx, params)
	pool := keeper.GetPool(ctx)

	// create some candidates some bonded, some unbonded. the largest candidates are bonded
	// so that we can have a big change in bonded ratio (~55 to ~73) when cand9 bonds
	candidates := make([]Candidate, 10)
	for i := 0; i < 10; i++ {
		c := Candidate{
			Status:      Unbonded,
			PubKey:      pks[i],
			Address:     addrs[i],
			Assets:      sdk.NewRat(0),
			Liabilities: sdk.NewRat(0),
		}

		//leave candidate 9 unbonded, so we can bond it later
		if i > 4 && i < 9 {
			c.Status = Bonded
		}
		mintedTokens := int64((i + 1) * 10000000)
		pool.TotalSupply += mintedTokens
		pool, c, _ = pool.candidateAddTokens(c, mintedTokens)
		keeper.setCandidate(ctx, c)
		candidates[i] = c
	}
	keeper.setPool(ctx, pool)

	//initialzing variables for candidate state. These will get updated throughout the test
	var totalSupply, bondedTokens, unbondedTokens, cand9unbondedTokens, cand9bondedTokens int64 = 550000000, 300000000, 250000000, 100000000, 0
	//share to token ratio starts off as 1:1
	var bondedShares, unbondedShares, unbondSharesCand9 sdk.Rat = sdk.NewRat(300000000, 1), sdk.NewRat(250000000, 1), sdk.NewRat(100000000, 1)
	// cumulative count of provisions, so we can check at the end of the year that the bonded pool has increased by this amount
	var cumulativeExpProvs int64

	assert.Equal(t, totalSupply, pool.TotalSupply)
	assert.Equal(t, bondedTokens, pool.BondedPool)
	assert.Equal(t, unbondedTokens, pool.UnbondedPool)

	// initial bonded ratio ~ 54%
	assert.True(t, pool.bondedRatio().Equal(sdk.NewRat(bondedTokens, totalSupply)), "%v", pool.bondedRatio())

	// test the value of candidate shares. should start off 1:1
	assert.True(t, pool.bondedShareExRate().Equal(sdk.OneRat()), "%v", pool.bondedShareExRate())

	initialTotalTokens := pool.TotalSupply
	initialUnbondedTokens := pool.TotalSupply - pool.BondedPool

	// process the provisions a year
	for hr := 0; hr < 8766; hr++ {

		pool := keeper.GetPool(ctx)
		_, expProvisions, pool := checkHourlyProvisions(t, keeper, pool, ctx, hr)
		cumulativeExpProvs = cumulativeExpProvs + expProvisions

		//hour 1600 was arbitrarily picked to bond the largest candidate
		if hr <= 1600 {
			require.Equal(t, unbondedTokens, pool.UnbondedPool)
		} else {
			require.Equal(t, initialUnbondedTokens-cand9bondedTokens, pool.UnbondedPool)
		}

		//steps to create unbonding of a candidate at hour 1600, to increase ratio from ~55% to ~73%
		if hr == 1600 {
			candidate, found := keeper.GetCandidate(ctx, addrs[9])
			assert.True(t, found)
			beforeBondedRatio := pool.bondedRatio()

			// bond 100,000,000 tokens that were previously unbonded
			pool, candidate = pool.unbondedToBondedPool(candidate)
			keeper.setPool(ctx, pool)
			unbondedShares = unbondedShares.Sub(unbondSharesCand9)

			//candidate.Assets are the shares. shares should be less than 100,000, bondedShareExRate() should be greater than 1, when multiplied they equal 100,000
			cand9bondedTokens = pool.bondedShareExRate().Mul(candidate.Assets).Evaluate()
			cand9unbondedTokens = cand9unbondedTokens - cand9bondedTokens
			assert.Equal(t, int64(100000000), cand9bondedTokens)

			// must add cumulativeExpProvs here, to get true bonded tokens at this instance, to find new value for bondedShares
			bondedAt1600 := bondedTokens + cand9bondedTokens + cumulativeExpProvs

			// bonded shares should increase
			bondedShares = sdk.NewRat(bondedAt1600, 1).Quo(pool.bondedShareExRate())
			assert.True(t, bondedShares.GT(sdk.NewRat(300000000, 1)))

			//Ensure that new bonded ratio is greater than old bonded ratio, since we just added 100,000 bonded
			assert.True(t, (pool.bondedRatio().GT(beforeBondedRatio)))
		}
	}

	// Final check that the pool equals initial values + provisions and adjustments we recorded
	pool = keeper.GetPool(ctx)
	checkFinalPoolValues(t, pool, initialTotalTokens,
		initialUnbondedTokens, cumulativeExpProvs,
		cand9bondedTokens, -cand9bondedTokens, bondedShares, unbondedShares)

}

//Tests that inflation works as expected when we get a randomly updating sample of candidates
func TestAddingRandomCandidates(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)
	params := defaultParams()
	keeper.setParams(ctx, params)
	r := rand.New(rand.NewSource(502))
	numCandidates := 20 //max 40 right now since var addrs only goes up to addrs[39]

	//start off by randomly creating 20 candidates
	pool, candidates := randomSetup(r, numCandidates)
	require.Equal(t, numCandidates, len(candidates))

	for i := 0; i < len(candidates); i++ {
		keeper.setCandidate(ctx, candidates[i])
	}

	keeper.setPool(ctx, pool)

	//Every two weeks (336 hours) we do a random operation on the set of 20 candidates.
	twoWeekCounter := 336

	//We count up to the 20 total candidates, and do a random operation on each candidate
	candidateCounter := 0

	// One year of provisions, with 20 random operations from the 20 candidates setup above from randomSetup()
	for hr := 0; hr < 8766; hr++ {
		pool := keeper.GetPool(ctx)

		// This if statement will randomly bond, unbond, remove shares or add tokens to the candidates every two weeks, for 40 weeks
		// every other hour it will just add normal provisions
		if twoWeekCounter == hr && candidateCounter < 20 {

			//Get values before randomOperation
			expInflationBefore := keeper.nextInflation(ctx)
			initialBondedPool := pool.BondedPool
			initialUnbondedTokens := pool.UnbondedPool

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

			//Get values after randomOperation
			expInflationAfter := keeper.nextInflation(ctx)
			afterBondedPool := pool.BondedPool
			afterUnbondedPool := pool.UnbondedPool

			//Process provisions after random operation
			pool = keeper.processProvisions(ctx)
			keeper.setPool(ctx, pool)

			//Comparing expected inflation before and after the randomOperation()
			inflationIncreased := expInflationAfter.GT(expInflationBefore)
			inflationEqual := expInflationAfter.Equal(expInflationBefore)

			switch {
			//Inflation will NOT increase, because we are adding bonded tokens to the pool
			//CASE: Happens when we randomly add tokens to a bonded candidate
			//CASE: Happens when we randomly bond a candidate from unbonded
			case afterBondedPool > initialBondedPool:

				//for the off case where we are bonded so low (i.e. 15%) , that we add bonded tokens and inflation is unchanged at 20%
				//for the off case where we are bonded so high (i.e. 90%), that we add bonded tokens and inflation is unchanged at 7%
				if expInflationBefore.Equal(sdk.NewRat(20, 100)) {
					assert.True(t, !inflationIncreased || inflationEqual, msg)
				} else if expInflationBefore.Equal(sdk.NewRat(7, 100)) {
					assert.True(t, !inflationIncreased || inflationEqual, msg)
				} else {
					assert.True(t, !inflationIncreased, msg)
				}

			//Inflation WILL increase, because we are removing bonded tokens from the pool
			//CASE: Happens when we randomly remove bonded Shares
			//CASE: Happens when we randomly unbond a candidate that was bonded
			case afterBondedPool < initialBondedPool:

				//For the off case where we are bonded so low (i.e. 15%),  and more tokens are unbonded, inflation is unchanged at 20%
				//For the off case where we are bonded so high (i.e. 90%) and we unbond a small amount (maybe 2%), inflation  is unchanged at 7%
				if expInflationBefore.Equal(sdk.NewRat(20, 100)) {
					assert.True(t, inflationIncreased || inflationEqual, msg)
				} else if expInflationBefore.Equal(sdk.NewRat(7, 100)) {
					assert.True(t, inflationIncreased || inflationEqual, msg)
				} else {
					assert.True(t, inflationIncreased, msg)
				}

			//Inflation will STAY THE SAME.
			//Inflation is dependant only on bondedRatio. Sure, a validator can add unbonded tokens, but it doesn't change bondedRatio (totalBondedTokens / globalTotalTokens)
			//CASE: Happens when we randomly add unbonded tokens
			case afterUnbondedPool > initialUnbondedTokens:
				assert.True(t, inflationEqual, msg)

			//Inflation will STAY THE SAME.
			//Inflation is dependant only on bondedRatio. Sure, a validator can add unbonded tokens, but it doesn't change bondedRatio (totalBondedTokens / globalTotalTokens)
			//CASE: Happens when we randomly remove shares from an unbonded candidate
			case afterUnbondedPool < initialUnbondedTokens:
				assert.True(t, inflationEqual, msg)

			default:
				panic(fmt.Sprintf("pool.UnbondedPool and pool.BondedPool are unchanged. All operations should change either the unbondedPool or bondedPool amounts."))
			}

			twoWeekCounter += 336
			candidateCounter++

		} else {

			// If we are not doing a random operation, just check that normal provisions are working for each hour
			checkHourlyProvisions(t, keeper, pool, ctx, hr)

		}
	}
}

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
func checkHourlyProvisions(t *testing.T, keeper Keeper, pool Pool, ctx sdk.Context, hr int) (sdk.Rat, int64, Pool) {

	//If we are not doing a random operation, just check that normal provisions are working for each hour
	expInflation := keeper.nextInflation(ctx)
	expProvisions := (expInflation.Mul(sdk.NewRat(pool.TotalSupply)).Quo(hrsPerYrRat)).Evaluate()
	startBondedPool := pool.BondedPool
	startTotalSupply := pool.TotalSupply
	pool = keeper.processProvisions(ctx)
	keeper.setPool(ctx, pool)

	//check provisions were added to pool
	require.Equal(t, startBondedPool+expProvisions, pool.BondedPool, "hr %v", hr)
	require.Equal(t, startTotalSupply+expProvisions, pool.TotalSupply)

	return expInflation, expProvisions, pool
}
