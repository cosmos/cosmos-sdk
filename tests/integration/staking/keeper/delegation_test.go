package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"gotest.tools/v3/assert"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

func TestUnbondingDelegationsMaxEntries(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx

	addrs := simtestutil.AddTestAddrsIncremental(f.bankKeeper, f.stakingKeeper, ctx, 5, f.stakingKeeper.TokensFromConsensusPower(ctx, 100))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addrs)
	msgServer := keeper.NewMsgServerImpl(f.stakingKeeper)

	comm := types.NewCommissionRates(math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0))

	stakedTokens := f.stakingKeeper.TokensFromConsensusPower(ctx, 30)

	// create a validator with self delegation
	msg, err := types.NewMsgCreateValidator(valAddrs[0], PKs[0], sdk.NewCoin(sdk.DefaultBondDenom, stakedTokens),
		types.Description{Moniker: "NewVal"}, comm, math.OneInt())
	require.NoError(t, err)
	_, err = msgServer.CreateValidator(ctx, msg)
	require.NoError(t, err)

	_, err = f.stakingKeeper.EndBlocker(ctx)
	require.NoError(t, err)

	addrDel := sdk.AccAddress(valAddrs[0])
	addrVal := sdk.ValAddress(addrDel)

	bondDenom := f.stakingKeeper.BondDenom(ctx)
	maxEntries := f.stakingKeeper.MaxEntries(ctx)

	oldBonded := f.bankKeeper.GetBalance(ctx, f.stakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	oldNotBonded := f.bankKeeper.GetBalance(ctx, f.stakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	// should all pass
	var completionTime time.Time
	totalUnbonded := math.NewInt(0)
	for i := int64(0); i < int64(maxEntries); i++ {
		var err error
		ctx = ctx.WithBlockHeight(i)
		var amount math.Int
		completionTime, amount, err = f.stakingKeeper.Undelegate(ctx, addrDel, addrVal, math.LegacyNewDec(1))
		if totalUnbonded.IsZero() {
			totalUnbonded = amount
		} else {
			totalUnbonded = totalUnbonded.Add(amount)
		}

		assert.NilError(t, err)
	}

	newBonded := f.bankKeeper.GetBalance(ctx, f.stakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	newNotBonded := f.bankKeeper.GetBalance(ctx, f.stakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, newBonded, oldBonded.SubRaw(int64(maxEntries))))
	assert.Assert(math.IntEq(t, newNotBonded, oldNotBonded.AddRaw(int64(maxEntries))))
	assert.Assert(math.IntEq(t, totalUnbonded, oldBonded.Sub(newBonded)))
	assert.Assert(math.IntEq(t, totalUnbonded, newNotBonded.Sub(oldNotBonded)))

	oldBonded = f.bankKeeper.GetBalance(ctx, f.stakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	oldNotBonded = f.bankKeeper.GetBalance(ctx, f.stakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	// an additional unbond should fail due to max entries
	_, _, err = f.stakingKeeper.Undelegate(ctx, addrDel, addrVal, math.LegacyNewDec(1))
	assert.Error(t, err, "too many unbonding delegation entries for (delegator, validator) tuple")

	newBonded = f.bankKeeper.GetBalance(ctx, f.stakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	newNotBonded = f.bankKeeper.GetBalance(ctx, f.stakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	assert.Assert(math.IntEq(t, newBonded, oldBonded))
	assert.Assert(math.IntEq(t, newNotBonded, oldNotBonded))

	// mature unbonding delegations
	ctx = ctx.WithBlockTime(completionTime)
	_, err = f.stakingKeeper.CompleteUnbonding(ctx, addrDel, addrVal)
	assert.NilError(t, err)

	newBonded = f.bankKeeper.GetBalance(ctx, f.stakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	newNotBonded = f.bankKeeper.GetBalance(ctx, f.stakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, newBonded, oldBonded))
	assert.Assert(math.IntEq(t, newNotBonded, oldNotBonded.SubRaw(int64(maxEntries))))

	oldNotBonded = f.bankKeeper.GetBalance(ctx, f.stakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	// unbonding  should work again
	_, _, err = f.stakingKeeper.Undelegate(ctx, addrDel, addrVal, math.LegacyNewDec(1))
	assert.NilError(t, err)

	newBonded = f.bankKeeper.GetBalance(ctx, f.stakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	newNotBonded = f.bankKeeper.GetBalance(ctx, f.stakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount
	assert.Assert(math.IntEq(t, newBonded, oldBonded.SubRaw(1)))
	assert.Assert(math.IntEq(t, newNotBonded, oldNotBonded.AddRaw(1)))
}
