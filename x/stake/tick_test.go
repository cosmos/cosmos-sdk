package stake

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tmlibs/rational"
)

func TestGetInflation(t *testing.T) {
	store := initTestStore(t)
	params := loadParams(store)
	gs := loadGlobalState(store)

	// Governing Mechanism:
	//    bondedRatio = BondedPool / TotalSupply
	//    inflationRateChangePerYear = (1- bondedRatio/ GoalBonded) * MaxInflationRateChange

	tests := []struct {
		setBondedPool, setTotalSupply int64
		setInflation, expectedChange  rational.Rat
	}{
		// with 0% bonded atom supply the inflation should increase by InflationRateChange
		{0, 0, rational.New(7, 100), params.InflationRateChange.Quo(hrsPerYr)},

		// 100% bonded, starting at 20% inflation and being reduced
		{1, 1, rational.New(20, 100), rational.One.Sub(rational.One.Quo(params.GoalBonded)).Mul(params.InflationRateChange).Quo(hrsPerYr)},

		// 50% bonded, starting at 10% inflation and being increased
		{1, 2, rational.New(10, 100), rational.One.Sub(rational.New(1, 2).Quo(params.GoalBonded)).Mul(params.InflationRateChange).Quo(hrsPerYr)},

		// test 7% minimum stop (testing with 100% bonded)
		{1, 1, rational.New(7, 100), rational.Zero},
		{1, 1, rational.New(70001, 1000000), rational.New(-1, 1000000)},

		// test 20% maximum stop (testing with 0% bonded)
		{0, 0, rational.New(20, 100), rational.Zero},
		{0, 0, rational.New(199999, 1000000), rational.New(1, 1000000)},

		// perfect balance shouldn't change inflation
		{67, 100, rational.New(15, 100), rational.Zero},
	}
	for _, tc := range tests {
		gs.BondedPool, gs.TotalSupply = tc.setBondedPool, tc.setTotalSupply
		gs.Inflation = tc.setInflation

		inflation := nextInflation(gs, params)
		diffInflation := inflation.Sub(tc.setInflation)

		assert.True(t, diffInflation.Equal(tc.expectedChange),
			"%v, %v", diffInflation, tc.expectedChange)
	}
}

func TestProcessProvisions(t *testing.T) {
	store := initTestStore(t)
	params := loadParams(store)
	gs := loadGlobalState(store)

	// create some candidates some bonded, some unbonded
	n := 10
	actors := newActors(n)
	candidates := candidatesFromActorsEmpty(actors)
	for i, candidate := range candidates {
		if i < 5 {
			candidate.Status = Bonded
		}
		mintedTokens := int64((i + 1) * 10000000)
		gs.TotalSupply += mintedTokens
		candidate.addTokens(mintedTokens, gs)
		saveCandidate(store, candidate)
	}
	var totalSupply int64 = 550000000
	var bondedShares int64 = 150000000
	var unbondedShares int64 = 400000000

	// initial bonded ratio ~ 27%
	assert.True(t, gs.bondedRatio().Equal(rational.New(bondedShares, totalSupply)), "%v", gs.bondedRatio())

	// Supplies
	assert.Equal(t, totalSupply, gs.TotalSupply)
	assert.Equal(t, bondedShares, gs.BondedPool)
	assert.Equal(t, unbondedShares, gs.UnbondedPool)

	// test the value of candidate shares
	assert.True(t, gs.bondedShareExRate().Equal(rational.One), "%v", gs.bondedShareExRate())

	initialSupply := gs.TotalSupply
	initialUnbonded := gs.TotalSupply - gs.BondedPool

	// process the provisions a year
	for hr := 0; hr < 8766; hr++ {
		expInflation := nextInflation(gs, params).Round(1000000000)
		expProvisions := (expInflation.Mul(rational.New(gs.TotalSupply)).Quo(hrsPerYr)).Evaluate()
		startBondedPool := gs.BondedPool
		startTotalSupply := gs.TotalSupply
		processProvisions(store, gs, params)
		assert.Equal(t, startBondedPool+expProvisions, gs.BondedPool)
		assert.Equal(t, startTotalSupply+expProvisions, gs.TotalSupply)
	}
	assert.NotEqual(t, initialSupply, gs.TotalSupply)
	assert.Equal(t, initialUnbonded, gs.UnbondedPool)
	//panic(fmt.Sprintf("debug total %v, bonded  %v, diff %v\n", gs.TotalSupply, gs.BondedPool, gs.TotalSupply-gs.BondedPool))

	// initial bonded ratio ~ 35% ~ 30% increase for bonded holders
	assert.True(t, gs.bondedRatio().Equal(rational.New(105906511, 305906511)), "%v", gs.bondedRatio())

	// global supply
	assert.Equal(t, int64(611813022), gs.TotalSupply)
	assert.Equal(t, int64(211813022), gs.BondedPool)
	assert.Equal(t, unbondedShares, gs.UnbondedPool)

	// test the value of candidate shares
	assert.True(t, gs.bondedShareExRate().Mul(rational.New(bondedShares)).Equal(rational.New(211813022)), "%v", gs.bondedShareExRate())

}
