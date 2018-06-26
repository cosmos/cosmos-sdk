package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
)

// setup helper function
// creates two validators
func setupHelper(t *testing.T) (sdk.Context, Keeper, types.Params, sdk.Address, crypto.PubKey) {
	// setup
	ctx, _, keeper := CreateTestInput(t, false, 10)
	amt := int64(10)
	addr := addrVals[0]
	pk := PKs[0]
	params := keeper.GetParams(ctx)
	pool := keeper.GetPool(ctx)
	pool.LooseTokens = 20

	// add a validator
	validator := types.NewValidator(addr, pk, types.Description{})
	validator, pool, _ = validator.AddTokensFromDel(pool, amt)
	keeper.SetPool(ctx, pool)
	keeper.UpdateValidator(ctx, validator)
	keeper.SetValidatorByPubKeyIndex(ctx, validator)

	// add a second validator
	validator = types.NewValidator(addrVals[1], PKs[1], types.Description{})
	validator, pool, _ = validator.AddTokensFromDel(pool, amt)
	keeper.SetPool(ctx, pool)
	keeper.UpdateValidator(ctx, validator)
	keeper.SetValidatorByPubKeyIndex(ctx, validator)

	return ctx, keeper, params, addr, pk
}

// tests Revoke, Unrevoke
func TestRevocation(t *testing.T) {
	// setup
	ctx, keeper, _, addr, pk := setupHelper(t)

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
	ctx, keeper, params, _, _ := setupHelper(t)
	fraction := sdk.NewRat(1).Quo(sdk.NewRat(2))

	// add an unbonding delegation past the current height
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
	ctx, keeper, params, _, _ := setupHelper(t)
	fraction := sdk.NewRat(1).Quo(sdk.NewRat(2))

	del := types.Delegation{
		DelegatorAddr: addrDels[0],
		ValidatorAddr: addrVals[1],
		Shares:        sdk.NewRat(10),
	}
	keeper.SetDelegation(ctx, del)

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

	// prior to the current height, stake didn't contribute
	slashAmount := keeper.slashRedelegation(ctx, rd, 1, fraction)
	require.Equal(t, int64(0), slashAmount.Evaluate())

	// after the expiration time, no longer eligible for slashing
	ctx = ctx.WithBlockHeader(abci.Header{Time: int64(10)})
	keeper.SetRedelegation(ctx, rd)
	slashAmount = keeper.slashRedelegation(ctx, rd, 0, fraction)
	require.Equal(t, int64(0), slashAmount.Evaluate())

	ctx = ctx.WithBlockHeader(abci.Header{Time: int64(0)})
	keeper.SetRedelegation(ctx, rd)
	slashAmount = keeper.slashRedelegation(ctx, rd, 0, fraction)
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

// tests Slash at the current height
func TestSlashAtCurrentHeight(t *testing.T) {
}

// tests Slash at a previous height with an unbonding delegation
func TestSlashWithUnbondingDelegation(t *testing.T) {
}

// tests Slash at a previous height with a redelegation
func TestSlashWithRedelegation(t *testing.T) {
}

// tests Slash at a previous height with a combination of unbonding delegations and redelegations
func TestSlashComplex(t *testing.T) {
}
