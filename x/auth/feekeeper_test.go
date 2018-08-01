package auth

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/params"
		)

var (
	emptyCoins = sdk.Coins{}
	oneCoin    = sdk.Coins{sdk.NewCoin("foocoin", 1)}
	twoCoins   = sdk.Coins{sdk.NewCoin("foocoin", 2)}
)

func TestFeeCollectionKeeperGetSet(t *testing.T) {
	ms, _, capKey2, _ := setupMultiStore()
	cdc := wire.NewCodec()
	paramKeeper := params.NewKeeper(cdc, sdk.NewKVStoreKey("params"))

	// make context and keeper
	ctx := sdk.NewContext(ms, abci.Header{}, false, log.NewNopLogger())
	fck := NewFeeCollectionKeeper(cdc, capKey2, paramKeeper.Getter())

	// no coins initially
	currFees := fck.GetCollectedFees(ctx)
	require.True(t, currFees.IsEqual(emptyCoins))

	// set feeCollection to oneCoin
	fck.setCollectedFees(ctx, oneCoin)

	// check that it is equal to oneCoin
	require.True(t, fck.GetCollectedFees(ctx).IsEqual(oneCoin))
}

func TestFeeCollectionKeeperAdd(t *testing.T) {
	ms, _, capKey2, _ := setupMultiStore()
	cdc := wire.NewCodec()
	paramKeeper := params.NewKeeper(cdc, sdk.NewKVStoreKey("params"))

	// make context and keeper
	ctx := sdk.NewContext(ms, abci.Header{}, false, log.NewNopLogger())
	fck := NewFeeCollectionKeeper(cdc, capKey2, paramKeeper.Getter())

	// no coins initially
	require.True(t, fck.GetCollectedFees(ctx).IsEqual(emptyCoins))

	// add oneCoin and check that pool is now oneCoin
	fck.addCollectedFees(ctx, oneCoin)
	require.True(t, fck.GetCollectedFees(ctx).IsEqual(oneCoin))

	// add oneCoin again and check that pool is now twoCoins
	fck.addCollectedFees(ctx, oneCoin)
	require.True(t, fck.GetCollectedFees(ctx).IsEqual(twoCoins))
}

func TestFeeCollectionKeeperClear(t *testing.T) {
	ms, _, capKey2, _ := setupMultiStore()
	cdc := wire.NewCodec()
	paramKeeper := params.NewKeeper(cdc, sdk.NewKVStoreKey("params"))

	// make context and keeper
	ctx := sdk.NewContext(ms, abci.Header{}, false, log.NewNopLogger())
	fck := NewFeeCollectionKeeper(cdc, capKey2, paramKeeper.Getter())

	// set coins initially
	fck.setCollectedFees(ctx, twoCoins)
	require.True(t, fck.GetCollectedFees(ctx).IsEqual(twoCoins))

	// clear fees and see that pool is now empty
	fck.ClearCollectedFees(ctx)
	require.True(t, fck.GetCollectedFees(ctx).IsEqual(emptyCoins))
}

func TestFeeCollectionKeeperPreprocess(t *testing.T) {
	ms, _, capKey2, paramsKey := setupMultiStore()

	cdc := wire.NewCodec()
	paramKeeper := params.NewKeeper(cdc, paramsKey)

	// make context and keeper
	ctx := sdk.NewContext(ms, abci.Header{}, false, log.NewNopLogger())
	fck := NewFeeCollectionKeeper(cdc, capKey2, paramKeeper.Getter())
	InitGenesis(ctx, paramKeeper.Setter(), DefaultGenesisState())

	var err sdk.Error
	err = fck.FeePreprocess(ctx, oneCoin)
	require.Error(t,err,"")

	fee1 := sdk.Coins{sdk.NewCoin("iGas", 50)}
	err = fck.FeePreprocess(ctx, fee1)
	require.Error(t,err,"")

	fee2 := sdk.Coins{sdk.NewCoin("iGas", 100)}
	err = fck.FeePreprocess(ctx, fee2)
	require.NoError(t,err,"")

	fee3 := sdk.Coins{sdk.NewCoin("iris", 1)}
	err = fck.FeePreprocess(ctx, fee3)
	require.Error(t,err,"")

	exchangeRate := sdk.NewRatFromInt(sdk.NewInt(100), sdk.OneInt())
	exchangeRateBytes,errMsg := fck.cdc.MarshalBinary(exchangeRate)
	require.NoError(t, errMsg)
	paramKeeper.Setter().SetRaw(ctx, FeeExchangeRatePrefix+"iris", exchangeRateBytes)

	err = fck.FeePreprocess(ctx, fee3)
	require.NoError(t,err,"")
}
