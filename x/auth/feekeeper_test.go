package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/tmlibs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"
)

var (
	emptyCoins = sdk.Coins{}
	oneCoin    = sdk.Coins{sdk.NewCoin("foocoin", 1)}
	twoCoins   = sdk.Coins{sdk.NewCoin("foocoin", 2)}
)

func TestFeeCollectionKeeperGetSet(t *testing.T) {
	ms, _, capKey2 := setupMultiStore()
	cdc := wire.NewCodec()

	// make context and keeper
	ctx := sdk.NewContext(ms, abci.Header{}, false, log.NewNopLogger())
	fck := NewFeeCollectionKeeper(cdc, capKey2)

	// no coins initially
	currFees := fck.GetCollectedFees(ctx)
	assert.True(t, currFees.IsEqual(emptyCoins))

	// set feeCollection to oneCoin
	fck.setCollectedFees(ctx, oneCoin)

	// check that it is equal to oneCoin
	assert.True(t, fck.GetCollectedFees(ctx).IsEqual(oneCoin))
}

func TestFeeCollectionKeeperAdd(t *testing.T) {
	ms, _, capKey2 := setupMultiStore()
	cdc := wire.NewCodec()

	// make context and keeper
	ctx := sdk.NewContext(ms, abci.Header{}, false, log.NewNopLogger())
	fck := NewFeeCollectionKeeper(cdc, capKey2)

	// no coins initially
	assert.True(t, fck.GetCollectedFees(ctx).IsEqual(emptyCoins))

	// add oneCoin and check that pool is now oneCoin
	fck.addCollectedFees(ctx, oneCoin)
	assert.True(t, fck.GetCollectedFees(ctx).IsEqual(oneCoin))

	// add oneCoin again and check that pool is now twoCoins
	fck.addCollectedFees(ctx, oneCoin)
	assert.True(t, fck.GetCollectedFees(ctx).IsEqual(twoCoins))
}

func TestFeeCollectionKeeperClear(t *testing.T) {
	ms, _, capKey2 := setupMultiStore()
	cdc := wire.NewCodec()

	// make context and keeper
	ctx := sdk.NewContext(ms, abci.Header{}, false, log.NewNopLogger())
	fck := NewFeeCollectionKeeper(cdc, capKey2)

	// set coins initially
	fck.setCollectedFees(ctx, twoCoins)
	assert.True(t, fck.GetCollectedFees(ctx).IsEqual(twoCoins))

	// clear fees and see that pool is now empty
	fck.ClearCollectedFees(ctx)
	assert.True(t, fck.GetCollectedFees(ctx).IsEqual(emptyCoins))
}
