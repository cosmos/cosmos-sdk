package keeper_test

import (
	"bytes"
	"testing"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttime "github.com/cometbft/cometbft/types/time"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"cosmossdk.io/collections"
	"cosmossdk.io/log/v2"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func TestMigrateLegacyBalances(t *testing.T) {
	key := storetypes.NewKVStoreKey(banktypes.StoreKey)
	oKey := storetypes.NewObjectStoreKey(banktypes.ObjectStoreKey)
	testCtx := testutil.DefaultContextWithObjectStore(t, key, storetypes.NewTransientStoreKey("transient_test"), oKey)
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: cmttime.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig()
	storeService := runtime.NewKVStoreService(key)

	ctrl := gomock.NewController(t)
	authKeeper := banktestutil.NewMockAccountKeeper(ctrl)
	authKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

	k := keeper.NewBaseKeeper(
		encCfg.Codec,
		storeService,
		authKeeper,
		nil,
		authtypes.NewModuleAddress("gov").String(),
		log.NewNopLogger(),
	)

	legacyAddr := sdk.AccAddress(bytes.Repeat([]byte{0x01}, 20))
	modernAddr := sdk.AccAddress(bytes.Repeat([]byte{0x02}, 20))
	legacyDenom := "ibc/A63965DEF1B5459FD58F19C5B1938244B016B075DE7C813C51AE9278CA8AF5B9"
	modernDenom := "uatom"

	require.NoError(t, k.Balances.Set(ctx, collections.Join(modernAddr, modernDenom), math.NewInt(5)))

	legacyCoin := sdk.NewInt64Coin(legacyDenom, 80000)
	rawValue, err := legacyCoin.Marshal()
	require.NoError(t, err)
	rawKey, err := collections.EncodeKeyWithPrefix(
		banktypes.BalancesPrefix.Bytes(),
		k.Balances.KeyCodec(),
		collections.Join(legacyAddr, legacyDenom),
	)
	require.NoError(t, err)
	require.NoError(t, storeService.OpenKVStore(ctx).Set(rawKey, rawValue))

	require.Equal(t, math.NewInt(80000), k.GetBalance(ctx, legacyAddr, legacyDenom).Amount)
	require.Empty(t, denomOwners(t, ctx, k, legacyDenom))

	n, err := k.MigrateLegacyBalances(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, n)

	owners := denomOwners(t, ctx, k, legacyDenom)
	require.Equal(t, []string{legacyAddr.String()}, owners)
	require.Equal(t, math.NewInt(80000), k.GetBalance(ctx, legacyAddr, legacyDenom).Amount)
	require.Equal(t, []string{modernAddr.String()}, denomOwners(t, ctx, k, modernDenom))
	require.Equal(t, math.NewInt(5), k.GetBalance(ctx, modernAddr, modernDenom).Amount)

	stored, err := storeService.OpenKVStore(ctx).Get(rawKey)
	require.NoError(t, err)
	gotAmt, err := sdk.IntValue.Decode(stored)
	require.NoError(t, err)
	require.Equal(t, math.NewInt(80000), gotAmt)

	n, err = k.MigrateLegacyBalances(ctx)
	require.NoError(t, err)
	require.Equal(t, 0, n)
}

func denomOwners(t *testing.T, ctx sdk.Context, k keeper.BaseKeeper, denom string) []string {
	t.Helper()
	iter, err := k.Balances.Indexes.Denom.MatchExact(ctx, denom)
	require.NoError(t, err)
	defer iter.Close()
	pks, err := iter.PrimaryKeys()
	require.NoError(t, err)
	out := make([]string, 0, len(pks))
	for _, pk := range pks {
		out = append(out, pk.K1().String())
	}
	return out
}
