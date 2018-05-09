package stake

import (
	"fmt"
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
	//    bondedRatio = BondedPool / TotalSupply
	//    inflationRateChangePerYear = (1- bondedRatio/ GoalBonded) * MaxInflationRateChange

	tests := []struct {
		name                          string
		setBondedPool, setTotalSupply int64
		setInflation, expectedChange  sdk.Rat
	}{
		// with 0% bonded atom supply the inflation should increase by InflationRateChange
		{"test 1", 0, 0, sdk.NewRat(7, 100), params.InflationRateChange.Quo(hrsPerYrRat).Round(precision)},

		// 100% bonded, starting at 20% inflation and being reduced
		// (1 - (1/0.67))*(0.13/8667)
		{"test 2", 1, 1, sdk.NewRat(20, 100),
			sdk.OneRat().Sub(sdk.OneRat().Quo(params.GoalBonded)).Mul(params.InflationRateChange).Quo(hrsPerYrRat).Round(precision)},

		// 50% bonded, starting at 10% inflation and being increased
		{"test 3", 1, 2, sdk.NewRat(10, 100),
			sdk.OneRat().Sub(sdk.NewRat(1, 2).Quo(params.GoalBonded)).Mul(params.InflationRateChange).Quo(hrsPerYrRat).Round(precision)},

		// test 7% minimum stop (testing with 100% bonded)
		{"test 4", 1, 1, sdk.NewRat(7, 100), sdk.ZeroRat()},
		{"test 5", 1, 1, sdk.NewRat(70001, 1000000), sdk.NewRat(-1, 1000000).Round(precision)},

		// test 20% maximum stop (testing with 0% bonded)
		{"test 6", 0, 0, sdk.NewRat(20, 100), sdk.ZeroRat()},
		{"test 7", 0, 0, sdk.NewRat(199999, 1000000), sdk.NewRat(1, 1000000).Round(precision)},

		// perfect balance shouldn't change inflation
		{"test 8", 67, 100, sdk.NewRat(15, 100), sdk.ZeroRat()},
	}
	for _, tc := range tests {
		pool.BondedPool, pool.TotalSupply = tc.setBondedPool, tc.setTotalSupply
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
	var bondedShares int64 = 150000000
	var unbondedShares int64 = 400000000
	var provisionTallied int64 = 0
	assert.Equal(t, totalSupply, pool.TotalSupply)
	assert.Equal(t, bondedShares, pool.BondedPool)
	assert.Equal(t, unbondedShares, pool.UnbondedPool)

	// fmt.Printf("pool bonded ratio: %v\n", pool)
	bondedRatioTest := pool.bondedRatio()
	fmt.Println("pool bonded ratio: ", bondedRatioTest) // is 150 mil / 550 mil, or 15/55, divisor 5 is 3/11

	// initial bonded ratio ~ 27%
	assert.True(t, pool.bondedRatio().Equal(sdk.NewRat(bondedShares, totalSupply)), "%v", pool.bondedRatio())

	bondedShareExRateTest := pool.bondedShareExRate()
	fmt.Println("pool bondedSharExTRate: ", bondedShareExRateTest)

	// test the value of candidate shares
	assert.True(t, pool.bondedShareExRate().Equal(sdk.OneRat()), "%v", pool.bondedShareExRate())

	initialSupply := pool.TotalSupply
	initialUnbonded := pool.TotalSupply - pool.BondedPool

	// process the provisions a year
	for hr := 0; hr < 8766; hr++ {
		pool := keeper.GetPool(ctx)
		expInflation := keeper.nextInflation(ctx).Round(1000000000) //so expInflation is the previous inflation in the previous hour, plus or minus the small change
		// fmt.Printf("Inflation Rate For each hour %v\n", expInflation)
		expProvisions := (expInflation.Mul(sdk.NewRat(pool.TotalSupply)).Quo(hrsPerYrRat)).Evaluate() // this is %inflation (7-14%) * 550,000,000 / 8766 . provisions are handed out atoms
		// fmt.Printf("Provisons For each hour %v\n", expProvisions)

		//so here we are getting the previous amoutns we want to keep, then we run processProvisions which actually adds atoms
		//then we set the pool, so now its updated
		//now we compare the old values to the new values plus the inlfation and provisions we got at expInlfation and expProvisions
		startBondedPool := pool.BondedPool
		startTotalSupply := pool.TotalSupply
		provisionTallied = provisionTallied + expProvisions
		// fmt.Printf("Provisons For this hour: %v NewProvisionTotal : %v\n", expProvisions, provisionTally)
		pool = keeper.processProvisions(ctx)
		keeper.setPool(ctx, pool)
		//fmt.Printf("hr %v, startBondedPool %v, expProvisions %v, pool.BondedPool %v\n", hr, startBondedPool, expProvisions, pool.BondedPool)
		require.Equal(t, startBondedPool+expProvisions, pool.BondedPool, "hr %v", hr)
		require.Equal(t, startTotalSupply+expProvisions, pool.TotalSupply)
	}
	pool = keeper.GetPool(ctx)
	assert.NotEqual(t, initialSupply, pool.TotalSupply) //TODO: this is lazy
	assert.Equal(t, initialUnbonded, pool.UnbondedPool)
	//panic(fmt.Sprintf("debug total %v, bonded  %v, diff %v\n", p.TotalSupply, p.BondedPool, pool.TotalSupply-pool.BondedPool))
	fmt.Printf("debug total %v, \nbonded  %v, \ndiff %v\n", pool.TotalSupply, pool.BondedPool, pool.TotalSupply-pool.BondedPool)
	// fmt.Println("provisionTally: ", provisionTally)

	calculatedTotalSupply := totalSupply + provisionTallied
	calculatedBondedSupply := bondedShares + provisionTallied
	// initial bonded ratio ~ from 27% to 40% increase for bonded holders ownership of total supply
	assert.True(t, pool.bondedRatio().Equal(sdk.NewRat(calculatedBondedSupply, calculatedTotalSupply)), "%v", pool.bondedRatio())

	// global supply
	assert.Equal(t, calculatedTotalSupply, pool.TotalSupply)
	assert.Equal(t, calculatedBondedSupply, pool.BondedPool)
	assert.Equal(t, unbondedShares, pool.UnbondedPool)

	// test the value of candidate shares
	fmt.Printf("what is the new ex rate?: %v ", pool.bondedShareExRate())
	assert.True(t, pool.bondedShareExRate().Mul(sdk.NewRat(bondedShares)).Equal(sdk.NewRat(calculatedBondedSupply)), "%v", pool.bondedShareExRate())
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
	var bondedShares int64 = 150000000
	var unbondedShares int64 = 400000000
	var provisionTallied int64 = 0
	assert.Equal(t, totalSupply, pool.TotalSupply)
	assert.Equal(t, bondedShares, pool.BondedPool)
	assert.Equal(t, unbondedShares, pool.UnbondedPool)

	// // fmt.Printf("pool bonded ratio: %v\n", pool)
	// bondedRatioTest := pool.bondedRatio()
	// fmt.Println("pool bonded ratio: ", bondedRatioTest) // is 150 mil / 550 mil, or 15/55, divisor 5 is 3/11

	// initial bonded ratio ~ 27%
	assert.True(t, pool.bondedRatio().Equal(sdk.NewRat(bondedShares, totalSupply)), "%v", pool.bondedRatio())

	// bondedShareExRateTest := pool.bondedShareExRate()
	// fmt.Println("pool bondedSharExTRate: ", bondedShareExRateTest)

	// test the value of candidate shares
	assert.True(t, pool.bondedShareExRate().Equal(sdk.OneRat()), "%v", pool.bondedShareExRate())

	initialSupply := pool.TotalSupply
	initialUnbonded := pool.TotalSupply - pool.BondedPool

	// ~11.4 years to go from 7%, up to 20%, back down to 7%
	for hr := 0; hr < 100000; hr++ {
		pool := keeper.GetPool(ctx)
		expInflation := keeper.nextInflation(ctx).Round(1000000000)
		// expInflationFloat, _ := expInflation.Float64()
		// fmt.Println("")
		// fmt.Printf("Yearly Inflation Rate + hour adjusted: %v\n", expInflationFloat*100)
		expProvisions := (expInflation.Mul(sdk.NewRat(pool.TotalSupply)).Quo(hrsPerYrRat)).Evaluate()
		// fmt.Printf("Provisons For each hour %v\n", expProvisions)

		// pbr := pool.bondedRatio()
		// poolBondedRatio, _ := pbr.Float64()
		// fmt.Println("Pool bonded Ratio: ", poolBondedRatio*100)
		// fmt.Println("HOUR: ", hr)

		startBondedPool := pool.BondedPool
		startTotalSupply := pool.TotalSupply
		provisionTallied = provisionTallied + expProvisions
		previousInflation := pool.Inflation
		// fmt.Printf("Provisons For this hour: %v NewProvisionTotal : %v\n", expProvisions, provisionTally)
		pool = keeper.processProvisions(ctx)
		keeper.setPool(ctx, pool)
		//fmt.Printf("hr %v, startBondedPool %v, expProvisions %v, pool.BondedPool %v\n", hr, startBondedPool, expProvisions, pool.BondedPool)
		require.Equal(t, startBondedPool+expProvisions, pool.BondedPool, "hr %v", hr)
		require.Equal(t, startTotalSupply+expProvisions, pool.TotalSupply)
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
	pool = keeper.GetPool(ctx)
	assert.NotEqual(t, initialSupply, pool.TotalSupply)
	assert.Equal(t, initialUnbonded, pool.UnbondedPool)
	//panic(fmt.Sprintf("debug total %v, bonded  %v, diff %v\n", p.TotalSupply, p.BondedPool, pool.TotalSupply-pool.BondedPool))
	fmt.Printf("debug total %v, \nbonded  %v, \ndiff %v\n", pool.TotalSupply, pool.BondedPool, pool.TotalSupply-pool.BondedPool)
	// fmt.Println("provisionTally: ", provisionTally)

	calculatedTotalSupply := totalSupply + provisionTallied
	calculatedBondedSupply := bondedShares + provisionTallied
	// initial bonded ratio ~ from 27% to 40% increase for bonded holders ownership of total supply
	assert.True(t, pool.bondedRatio().Equal(sdk.NewRat(calculatedBondedSupply, calculatedTotalSupply)), "%v", pool.bondedRatio())

	// global supply
	assert.Equal(t, calculatedTotalSupply, pool.TotalSupply)
	assert.Equal(t, calculatedBondedSupply, pool.BondedPool)
	assert.Equal(t, unbondedShares, pool.UnbondedPool)

	// test the value of candidate shares
	// fmt.Printf("what is the new ex rate?: %v ", pool.bondedShareExRate())
	assert.True(t, pool.bondedShareExRate().Mul(sdk.NewRat(bondedShares)).Equal(sdk.NewRat(calculatedBondedSupply)), "%v", pool.bondedShareExRate())
}
