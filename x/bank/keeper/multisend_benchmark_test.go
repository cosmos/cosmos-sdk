package keeper_test

import (
	"fmt"
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/store/v2"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func setupMultiSend(tb testing.TB, n int, distinct bool) (sdk.Context, keeper.BaseKeeper, *banktypes.MsgMultiSend) {
	tb.Helper()
	keys := storetypes.NewKVStoreKeys(authtypes.StoreKey, banktypes.StoreKey)
	db := dbm.NewMemDB()
	tb.Cleanup(func() { require.NoError(tb, db.Close()) })
	cms := store.NewCommitMultiStore(db, log.NewNopLogger())
	for _, key := range keys {
		cms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, nil)
	}
	require.NoError(tb, cms.LoadLatestVersion())
	ctx := sdk.NewContext(cms, cmtproto.Header{Time: time.Unix(1, 0)}, false, log.NewNopLogger())
	encCfg := moduletestutil.MakeTestEncodingConfig()
	authtypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	banktypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	authority := authtypes.NewModuleAddress("gov").String()
	ak := authkeeper.NewAccountKeeper(encCfg.Codec, runtime.NewKVStoreService(keys[authtypes.StoreKey]), authtypes.ProtoBaseAccount, nil, address.NewBech32Codec("cosmos"), "cosmos", authority)
	bk := keeper.NewBaseKeeper(encCfg.Codec, runtime.NewKVStoreService(keys[banktypes.StoreKey]), ak, nil, authority, log.NewNopLogger())
	require.NoError(tb, bk.SetParams(ctx, banktypes.DefaultParams()))
	sender := sdk.AccAddress([]byte("multisend-bench-input"))
	ak.SetAccount(ctx, ak.NewAccountWithAddress(ctx, sender))
	outputs := make([]banktypes.Output, n)
	coins := make(sdk.Coins, 0, n)
	for i := range outputs {
		receiver := sdk.AccAddress([]byte(fmt.Sprintf("multisend-output%04d", i)))
		ak.SetAccount(ctx, ak.NewAccountWithAddress(ctx, receiver))
		denom := "token"
		if distinct {
			denom = fmt.Sprintf("token%03d", i)
			coins = append(coins, sdk.NewInt64Coin(denom, 1))
		}
		outputs[i] = banktypes.NewOutput(receiver, sdk.Coins{sdk.NewInt64Coin(denom, 1)})
	}
	if !distinct {
		coins = sdk.Coins{sdk.NewInt64Coin("token", int64(n))}
	}
	for _, coin := range coins {
		require.NoError(tb, bk.Balances.Set(ctx, collections.Join(sender, coin.Denom), coin.Amount))
		require.NoError(tb, bk.Supply.Set(ctx, coin.Denom, coin.Amount))
	}
	cms.Commit()
	return ctx, bk, &banktypes.MsgMultiSend{Inputs: []banktypes.Input{banktypes.NewInput(sender, coins)}, Outputs: outputs}
}

func TestMultiSendManyDenominations(t *testing.T) {
	for _, distinct := range []bool{false, true} {
		for _, n := range []int{1, 10, 100} {
			t.Run(fmt.Sprintf("distinct=%t/N=%d", distinct, n), func(t *testing.T) {
				parent, bk, msg := setupMultiSend(t, n, distinct)
				ctx, _ := parent.CacheContext()
				ctx = ctx.WithGasMeter(storetypes.NewInfiniteGasMeter())
				_, err := keeper.NewMsgServerImpl(bk).MultiSend(ctx, msg)
				require.NoError(t, err)
				t.Logf("transfer gas: %d", ctx.GasMeter().GasConsumed())
				sender := sdk.MustAccAddressFromBech32(msg.Inputs[0].Address)
				require.Empty(t, bk.GetAllBalances(ctx, sender))
				for _, out := range msg.Outputs {
					receiver := sdk.MustAccAddressFromBech32(out.Address)
					require.Equal(t, out.Coins, bk.GetAllBalances(ctx, receiver))
					require.Contains(t, ctx.EventManager().Events(), sdk.NewEvent(banktypes.EventTypeTransfer,
						sdk.NewAttribute(banktypes.AttributeKeyRecipient, out.Address),
						sdk.NewAttribute(banktypes.AttributeKeySender, msg.Inputs[0].Address),
						sdk.NewAttribute(sdk.AttributeKeyAmount, out.Coins.String())))
				}
				for _, coin := range msg.Inputs[0].Coins {
					require.Equal(t, coin, bk.GetSupply(ctx, coin.Denom))
				}
			})
		}
	}
}

func BenchmarkMultiSend(b *testing.B) {
	for _, distinct := range []bool{false, true} {
		for _, n := range []int{1, 10, 100} {
			b.Run(fmt.Sprintf("distinct=%t/N=%d", distinct, n), func(b *testing.B) {
				parent, bk, msg := setupMultiSend(b, n, distinct)
				server := keeper.NewMsgServerImpl(bk)
				b.ReportAllocs()
				for b.Loop() {
					// Discard writes to repeat the same transfer from the same balances.
					ctx, _ := parent.CacheContext()
					if _, err := server.MultiSend(ctx, msg); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}
