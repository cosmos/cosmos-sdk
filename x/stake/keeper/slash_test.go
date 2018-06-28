package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
	abci "github.com/tendermint/abci/types"
)

// setup helper function
// creates two validators
func setupHelper(t *testing.T, amt int64) (sdk.Context, Keeper, types.Params) {
	// setup
	ctx, _, keeper := CreateTestInput(t, false, amt)
	params := keeper.GetParams(ctx)
	pool := keeper.GetPool(ctx)
	numVals := 3
	pool.LooseTokens = amt * int64(numVals)

	// add numVals validators
	for i := 0; i < numVals; i++ {
		validator := types.NewValidator(addrVals[i], PKs[i], types.Description{})
		validator, pool, _ = validator.AddTokensFromDel(pool, amt)
		keeper.SetPool(ctx, pool)
		keeper.UpdateValidator(ctx, validator)
		keeper.SetValidatorByPubKeyIndex(ctx, validator)
	}

	return ctx, keeper, params
}

// tests Revoke, Unrevoke
func TestRevocation(t *testing.T) {
	// setup
	ctx, keeper, _ := setupHelper(t, 10)
	addr := addrVals[0]
	pk := PKs[0]

	// initial state
	val, found := keeper.GetValidator(ctx, addr)
	require.True(t, found)
	require.False(t, val.GetRevoked())

	// test revoke
	keeper.Revoke(ctx, pk)
	val, found = keeper.GetValidator(ctx, addr)
	require.True(t, found)
	require.True(t, val.GetRevoked())

	// test unrevoke
	keeper.Unrevoke(ctx, pk)
	val, found = keeper.GetValidator(ctx, addr)
	require.True(t, found)
	require.False(t, val.GetRevoked())

}

// tests slashUnbondingDelegation
func TestSlashUnbondingDelegation(t *testing.T) {
	ctx, keeper, params := setupHelper(t, 10)
	fraction := sdk.NewRat(1, 2)

	// set an unbonding delegation
	ubd := types.UnbondingDelegation{
		DelegatorAddr:  addrDels[0],
		ValidatorAddr:  addrVals[0],
		CreationHeight: 0,
		MinTime:        0,
		InitialBalance: sdk.NewCoin(params.BondDenom, 10),
		Balance:        sdk.NewCoin(params.BondDenom, 10),
	}
	keeper.SetUnbondingDelegation(ctx, ubd)

	// prior to the current height, stake didn't contribute
	slashAmount := keeper.slashUnbondingDelegation(ctx, ubd, 1, fraction)
	require.Equal(t, int64(0), slashAmount.Evaluate())

	// after the expiration time, no longer eligible for slashing
	ctx = ctx.WithBlockHeader(abci.Header{Time: int64(10)})
	keeper.SetUnbondingDelegation(ctx, ubd)
	slashAmount = keeper.slashUnbondingDelegation(ctx, ubd, 0, fraction)
	require.Equal(t, int64(0), slashAmount.Evaluate())

	ctx = ctx.WithBlockHeader(abci.Header{Time: int64(0)})
	keeper.SetUnbondingDelegation(ctx, ubd)
	slashAmount = keeper.slashUnbondingDelegation(ctx, ubd, 0, fraction)
	require.Equal(t, int64(5), slashAmount.Evaluate())
	ubd, found := keeper.GetUnbondingDelegation(ctx, addrDels[0], addrVals[0])
	require.True(t, found)
	// initialbalance unchanged
	require.Equal(t, sdk.NewCoin(params.BondDenom, 10), ubd.InitialBalance)
	// balance decreased
	require.Equal(t, sdk.NewCoin(params.BondDenom, 5), ubd.Balance)
}

// tests slashRedelegation
func TestSlashRedelegation(t *testing.T) {
	ctx, keeper, params := setupHelper(t, 10)
	fraction := sdk.NewRat(1, 2)

	// set a redelegation
	rd := types.Redelegation{
		DelegatorAddr:    addrDels[0],
		ValidatorSrcAddr: addrVals[0],
		ValidatorDstAddr: addrVals[1],
		CreationHeight:   0,
		MinTime:          0,
		SharesSrc:        sdk.NewRat(10),
		SharesDst:        sdk.NewRat(10),
		InitialBalance:   sdk.NewCoin(params.BondDenom, 10),
		Balance:          sdk.NewCoin(params.BondDenom, 10),
	}
	keeper.SetRedelegation(ctx, rd)

	// set the associated delegation
	del := types.Delegation{
		DelegatorAddr: addrDels[0],
		ValidatorAddr: addrVals[1],
		Shares:        sdk.NewRat(10),
	}
	keeper.SetDelegation(ctx, del)

	// prior to the current height, stake didn't contribute
	slashAmount, _ := keeper.slashRedelegation(ctx, rd, 1, fraction)
	require.Equal(t, int64(0), slashAmount.Evaluate())

	// after the expiration time, no longer eligible for slashing
	ctx = ctx.WithBlockHeader(abci.Header{Time: int64(10)})
	keeper.SetRedelegation(ctx, rd)
	slashAmount, _ = keeper.slashRedelegation(ctx, rd, 0, fraction)
	require.Equal(t, int64(0), slashAmount.Evaluate())

	ctx = ctx.WithBlockHeader(abci.Header{Time: int64(0)})
	keeper.SetRedelegation(ctx, rd)
	slashAmount, _ = keeper.slashRedelegation(ctx, rd, 0, fraction)
	require.Equal(t, int64(5), slashAmount.Evaluate())
	rd, found := keeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.True(t, found)
	// initialbalance unchanged
	require.Equal(t, sdk.NewCoin(params.BondDenom, 10), rd.InitialBalance)
	// balance decreased
	require.Equal(t, sdk.NewCoin(params.BondDenom, 5), rd.Balance)

	// shares decreased
	del, found = keeper.GetDelegation(ctx, addrDels[0], addrVals[1])
	require.True(t, found)
	require.Equal(t, int64(5), del.Shares.Evaluate())
}

// tests Slash at a future height (must panic)
func TestSlashAtFutureHeight(t *testing.T) {
	ctx, keeper, _ := setupHelper(t, 10)
	pk := PKs[0]
	fraction := sdk.NewRat(1, 2)
	require.Panics(t, func() { keeper.Slash(ctx, pk, 1, 10, fraction) })
}

// tests Slash at the current height
func TestSlashAtCurrentHeight(t *testing.T) {
	ctx, keeper, _ := setupHelper(t, 10)
	pk := PKs[0]
	fraction := sdk.NewRat(1, 2)

	oldPool := keeper.GetPool(ctx)
	validator, found := keeper.GetValidatorByPubKey(ctx, pk)
	require.True(t, found)
	keeper.Slash(ctx, pk, ctx.BlockHeight(), 10, fraction)

	// read updated state
	validator, found = keeper.GetValidatorByPubKey(ctx, pk)
	require.True(t, found)
	newPool := keeper.GetPool(ctx)

	// power decreased
	require.Equal(t, sdk.NewRat(5), validator.GetPower())
	// pool bonded shares decreased
	require.Equal(t, sdk.NewRat(5).Evaluate(), oldPool.BondedShares.Sub(newPool.BondedShares).Evaluate())
}

// tests Slash at a previous height with an unbonding delegation
func TestSlashWithUnbondingDelegation(t *testing.T) {
	ctx, keeper, params := setupHelper(t, 10)
	pk := PKs[0]
	fraction := sdk.NewRat(1, 2)

	// set an unbonding delegation
	ubd := types.UnbondingDelegation{
		DelegatorAddr:  addrDels[0],
		ValidatorAddr:  addrVals[0],
		CreationHeight: 11,
		MinTime:        0,
		InitialBalance: sdk.NewCoin(params.BondDenom, 4),
		Balance:        sdk.NewCoin(params.BondDenom, 4),
	}
	keeper.SetUnbondingDelegation(ctx, ubd)

	// slash validator
	ctx = ctx.WithBlockHeight(12)
	oldPool := keeper.GetPool(ctx)
	validator, found := keeper.GetValidatorByPubKey(ctx, pk)
	require.True(t, found)
	keeper.Slash(ctx, pk, 10, 10, fraction)

	// read updating unbonding delegation
	ubd, found = keeper.GetUnbondingDelegation(ctx, addrDels[0], addrVals[0])
	require.True(t, found)
	// balance decreased
	require.Equal(t, sdk.NewInt(2), ubd.Balance.Amount)
	// read updated pool
	newPool := keeper.GetPool(ctx)
	// bonded tokens burned
	require.Equal(t, int64(3), oldPool.BondedTokens-newPool.BondedTokens)
	// read updated validator
	validator, found = keeper.GetValidatorByPubKey(ctx, pk)
	// power decreased, but not by quite half, stake was bonded since
	require.Equal(t, sdk.NewRat(7), validator.GetPower())
}

// tests Slash at a previous height with a redelegation
func TestSlashWithRedelegation(t *testing.T) {
	ctx, keeper, params := setupHelper(t, 10)
	pk := PKs[0]
	fraction := sdk.NewRat(1, 2)

	// set a redelegation
	rd := types.Redelegation{
		DelegatorAddr:    addrDels[0],
		ValidatorSrcAddr: addrVals[0],
		ValidatorDstAddr: addrVals[1],
		CreationHeight:   11,
		MinTime:          0,
		SharesSrc:        sdk.NewRat(6),
		SharesDst:        sdk.NewRat(6),
		InitialBalance:   sdk.NewCoin(params.BondDenom, 6),
		Balance:          sdk.NewCoin(params.BondDenom, 6),
	}
	keeper.SetRedelegation(ctx, rd)

	// set the associated delegation
	del := types.Delegation{
		DelegatorAddr: addrDels[0],
		ValidatorAddr: addrVals[1],
		Shares:        sdk.NewRat(6),
	}
	keeper.SetDelegation(ctx, del)

	// slash validator
	ctx = ctx.WithBlockHeight(12)
	oldPool := keeper.GetPool(ctx)
	validator, found := keeper.GetValidatorByPubKey(ctx, pk)
	require.True(t, found)
	keeper.Slash(ctx, pk, 10, 10, fraction)

	// read updating redelegation
	rd, found = keeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.True(t, found)
	// balance decreased
	require.Equal(t, sdk.NewInt(3), rd.Balance.Amount)
	// read updated pool
	newPool := keeper.GetPool(ctx)
	// bonded tokens burned
	require.Equal(t, int64(4), oldPool.BondedTokens-newPool.BondedTokens)
	// read updated validator
	validator, found = keeper.GetValidatorByPubKey(ctx, pk)
	// power decreased, but not by quite half, stake was bonded since
	require.Equal(t, sdk.NewRat(8), validator.GetPower())
}

// tests Slash at a previous height with both an unbonding delegation and a redelegation
func TestSlashBoth(t *testing.T) {
	ctx, keeper, params := setupHelper(t, 10)
	fraction := sdk.NewRat(1, 2)

	// set a redelegation
	rdA := types.Redelegation{
		DelegatorAddr:    addrDels[0],
		ValidatorSrcAddr: addrVals[0],
		ValidatorDstAddr: addrVals[1],
		CreationHeight:   11,
		MinTime:          0,
		SharesSrc:        sdk.NewRat(6),
		SharesDst:        sdk.NewRat(6),
		InitialBalance:   sdk.NewCoin(params.BondDenom, 6),
		Balance:          sdk.NewCoin(params.BondDenom, 6),
	}
	keeper.SetRedelegation(ctx, rdA)

	// set the associated delegation
	delA := types.Delegation{
		DelegatorAddr: addrDels[0],
		ValidatorAddr: addrVals[1],
		Shares:        sdk.NewRat(6),
	}
	keeper.SetDelegation(ctx, delA)

	// set an unbonding delegation
	ubdA := types.UnbondingDelegation{
		DelegatorAddr:  addrDels[0],
		ValidatorAddr:  addrVals[0],
		CreationHeight: 11,
		MinTime:        0,
		InitialBalance: sdk.NewCoin(params.BondDenom, 4),
		Balance:        sdk.NewCoin(params.BondDenom, 4),
	}
	keeper.SetUnbondingDelegation(ctx, ubdA)

	// slash validator
	ctx = ctx.WithBlockHeight(12)
	oldPool := keeper.GetPool(ctx)
	validator, found := keeper.GetValidatorByPubKey(ctx, PKs[0])
	require.True(t, found)
	keeper.Slash(ctx, PKs[0], 10, 10, fraction)

	// read updating redelegation
	rdA, found = keeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.True(t, found)
	// balance decreased
	require.Equal(t, sdk.NewInt(3), rdA.Balance.Amount)
	// read updated pool
	newPool := keeper.GetPool(ctx)
	// loose tokens burned
	require.Equal(t, int64(2), oldPool.LooseTokens-newPool.LooseTokens)
	// bonded tokens burned
	require.Equal(t, int64(3), oldPool.BondedTokens-newPool.BondedTokens)
	// read updated validator
	validator, found = keeper.GetValidatorByPubKey(ctx, PKs[0])
	// power not decreased, all stake was bonded since
	require.Equal(t, sdk.NewRat(10), validator.GetPower())
}
