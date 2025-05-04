package keeper_test

import (
	"testing"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttime "github.com/cometbft/cometbft/types/time"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func TestBankStateCompatibility(t *testing.T) {
	key := storetypes.NewKVStoreKey(banktypes.StoreKey)
	okey := storetypes.NewObjectStoreKey(banktypes.ObjectStoreKey)
	testCtx := testutil.DefaultContextWithObjectStore(t, key, storetypes.NewTransientStoreKey("transient_test"), okey)
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: cmttime.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig()
	storeService := runtime.NewKVStoreService(key)
	tkey := storetypes.NewTransientStoreKey(banktypes.TStoreKey)
	tStoreService := runtime.NewTransientKVStoreService(tkey)

	// gomock initializations
	ctrl := gomock.NewController(t)
	authKeeper := banktestutil.NewMockAccountKeeper(ctrl)
	authKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

	k := keeper.NewBaseKeeper(
		encCfg.Codec,
		storeService,
		tStoreService,
		okey,
		authKeeper,
		map[string]bool{accAddrs[4].String(): true},
		authtypes.NewModuleAddress("gov").String(),
		log.NewNopLogger(),
	)

	// test we can decode balances without problems
	// using the old value format of the denom to address index
	bankDenomAddressLegacyIndexValue := []byte{0} // taken from: https://github.com/cosmos/cosmos-sdk/blob/v0.47.3/x/bank/keeper/send.go#L361
	rawKey, err := collections.EncodeKeyWithPrefix(
		banktypes.DenomAddressPrefix,
		k.Balances.Indexes.Denom.KeyCodec(),
		collections.Join("atom", sdk.AccAddress("test")),
	)
	require.NoError(t, err)
	// we set the index key to the old value.
	require.NoError(t, storeService.OpenKVStore(ctx).Set(rawKey, bankDenomAddressLegacyIndexValue))

	// test walking is ok
	err = k.Balances.Indexes.Denom.Walk(ctx, nil, func(indexingKey string, indexedKey sdk.AccAddress) (stop bool, err error) {
		require.Equal(t, indexedKey, sdk.AccAddress("test"))
		require.Equal(t, indexingKey, "atom")
		return true, nil
	})
	require.NoError(t, err)

	// test matching is also ok
	iter, err := k.Balances.Indexes.Denom.MatchExact(ctx, "atom")
	require.NoError(t, err)
	defer iter.Close()
	pks, err := iter.PrimaryKeys()
	require.NoError(t, err)
	require.Len(t, pks, 1)
	require.Equal(t, pks[0], collections.Join(sdk.AccAddress("test"), "atom"))

	// assert the index value will be updated to the new format
	err = k.Balances.Indexes.Denom.Reference(ctx, collections.Join(sdk.AccAddress("test"), "atom"), math.ZeroInt(), nil)
	require.NoError(t, err)

	newRawValue, err := storeService.OpenKVStore(ctx).Get(rawKey)
	require.NoError(t, err)
	require.Equal(t, []byte{}, newRawValue)
}
