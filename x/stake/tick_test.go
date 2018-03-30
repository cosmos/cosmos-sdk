package stake

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

func TestGetInflation(t *testing.T) {
	ctx, _, keeper := createTestInput(t, nil, false, 0)
	params := defaultParams()
	keeper.setParams(ctx, params)
	gs := keeper.GetPool(ctx)

	// Governing Mechanism:
	//    bondedRatio = BondedPool / TotalSupply
	//    inflationRateChangePerYear = (1- bondedRatio/ GoalBonded) * MaxInflationRateChange

	tests := []struct {
		setBondedPool, setTotalSupply int64
		setInflation, expectedChange  sdk.Rat
	}{
		// with 0% bonded atom supply the inflation should increase by InflationRateChange
		{0, 0, sdk.NewRat(7, 100), params.InflationRateChange.Quo(hrsPerYrRat)},

		// 100% bonded, starting at 20% inflation and being reduced
		{1, 1, sdk.NewRat(20, 100), sdk.OneRat.Sub(sdk.OneRat.Quo(params.GoalBonded)).Mul(params.InflationRateChange).Quo(hrsPerYrRat)},

		// 50% bonded, starting at 10% inflation and being increased
		{1, 2, sdk.NewRat(10, 100), sdk.OneRat.Sub(sdk.NewRat(1, 2).Quo(params.GoalBonded)).Mul(params.InflationRateChange).Quo(hrsPerYrRat)},

		// test 7% minimum stop (testing with 100% bonded)
		{1, 1, sdk.NewRat(7, 100), sdk.ZeroRat},
		{1, 1, sdk.NewRat(70001, 1000000), sdk.NewRat(-1, 1000000)},

		// test 20% maximum stop (testing with 0% bonded)
		{0, 0, sdk.NewRat(20, 100), sdk.ZeroRat},
		{0, 0, sdk.NewRat(199999, 1000000), sdk.NewRat(1, 1000000)},

		// perfect balance shouldn't change inflation
		{67, 100, sdk.NewRat(15, 100), sdk.ZeroRat},
	}
	for _, tc := range tests {
		gs.BondedPool, gs.TotalSupply = tc.setBondedPool, tc.setTotalSupply
		gs.Inflation = tc.setInflation

		keeper.setPool(ctx, gs)

		inflation := keeper.nextInflation(ctx)
		diffInflation := inflation.Sub(tc.setInflation)

		assert.True(t, diffInflation.Equal(tc.expectedChange),
			"%v, %v", diffInflation, tc.expectedChange)
	}
}

func TestProcessProvisions(t *testing.T) {
	ctx, _, keeper := createTestInput(t, nil, false, 0)
	params := defaultParams()
	keeper.setParams(ctx, params)
	//gs := keeper.GetPool(ctx)

	totalSupply := int64(0)
	// create some candidates some bonded, some unbonded
	for i := 0; i < len(addrs); i++ {
		mintedTokens := int64((i + 1) * 10000000)
		c := Candidate{
			Status:      Unbonded,
			PubKey:      pks[i],
			Address:     addrs[i],
			Assets:      sdk.NewRat(mintedTokens),
			Liabilities: sdk.NewRat(mintedTokens),
		}
		if i < 5 {
			c.Status = Bonded
			keeper.addTokensBonded(ctx, mintedTokens)
		} else {
			keeper.addTokensUnbonded(ctx, mintedTokens)
		}
		totalSupply += mintedTokens
		keeper.candidateAddTokens(ctx, c, mintedTokens)
		keeper.setCandidate(ctx, c)
	}
	var bondedShares int64 = 150000000
	var unbondedShares int64 = 400000000

	gs := keeper.GetPool(ctx)
	gs.TotalSupply = totalSupply
	keeper.setPool(ctx, gs)

	// initial bonded ratio ~ 27%
	assert.True(t, gs.bondedRatio().Equal(sdk.NewRat(bondedShares, totalSupply)), "%v", gs.bondedRatio())

	// Supplies
	assert.Equal(t, totalSupply, gs.TotalSupply)
	assert.Equal(t, bondedShares, gs.BondedPool)
	assert.Equal(t, unbondedShares, gs.UnbondedPool)

	// test the value of candidate shares
	assert.True(t, gs.bondedShareExRate().Equal(sdk.OneRat), "%v", gs.bondedShareExRate())

	initialSupply := gs.TotalSupply
	initialUnbonded := gs.TotalSupply - gs.BondedPool

	// process the provisions a year
	for hr := 0; hr < 10; hr++ {
		fmt.Printf("start\n")
		expInflation := keeper.nextInflation(ctx).Round(1000000000)
		expProvisions := (expInflation.Mul(sdk.NewRat(gs.TotalSupply)).Quo(hrsPerYrRat)).Evaluate()

		fmt.Printf("gs:  %+v\n", gs)
		fmt.Printf("inf: %+v\n", expInflation)
		fmt.Printf("pro: %+v\n", expProvisions)

		startBondedPool := gs.BondedPool
		startTotalSupply := gs.TotalSupply
		keeper.processProvisions(ctx)
		gs = keeper.GetPool(ctx)
		assert.Equal(t, startBondedPool+expProvisions, gs.BondedPool)
		assert.Equal(t, startTotalSupply+expProvisions, gs.TotalSupply)
	}
	assert.NotEqual(t, initialSupply, gs.TotalSupply)
	assert.Equal(t, initialUnbonded, gs.UnbondedPool)
	//panic(fmt.Sprintf("debug total %v, bonded  %v, diff %v\n", p.TotalSupply, p.BondedPool, p.TotalSupply-gs.BondedPool))

	// initial bonded ratio ~ 35% ~ 30% increase for bonded holders
	assert.True(t, gs.bondedRatio().Equal(sdk.NewRat(105906511, 305906511)), "%v", gs.bondedRatio())

	// global supply
	assert.Equal(t, int64(611813022), gs.TotalSupply)
	assert.Equal(t, int64(211813022), gs.BondedPool)
	assert.Equal(t, unbondedShares, gs.UnbondedPool)

	// test the value of candidate shares
	assert.True(t, gs.bondedShareExRate().Mul(sdk.NewRat(bondedShares)).Equal(sdk.NewRat(211813022)), "%v", gs.bondedShareExRate())

}
