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
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/store/v2"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// Use committed IAVL stores and the real auth/bank keepers so the benchmark
// includes reading commission, saving the remainder and transferring coins.
func setupCommissionWithdrawal(tb testing.TB, n int) (sdk.Context, keeper.Keeper, bankkeeper.BaseKeeper, sdk.ValAddress) {
	tb.Helper()
	keys := storetypes.NewKVStoreKeys(authtypes.StoreKey, banktypes.StoreKey, types.StoreKey)
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
	ak := authkeeper.NewAccountKeeper(
		encCfg.Codec, runtime.NewKVStoreService(keys[authtypes.StoreKey]), authtypes.ProtoBaseAccount,
		map[string][]string{types.ModuleName: nil}, address.NewBech32Codec("cosmos"), "cosmos",
		authtypes.NewModuleAddress("gov").String(),
	)
	module := ak.GetModuleAccount(ctx, types.ModuleName)
	require.NotNil(tb, module)
	addr := sdk.AccAddress([]byte("commission-validator"))
	ak.SetAccount(ctx, ak.NewAccountWithAddress(ctx, addr))
	bk := bankkeeper.NewBaseKeeper(encCfg.Codec, runtime.NewKVStoreService(keys[banktypes.StoreKey]), ak, nil, authtypes.NewModuleAddress("gov").String(), log.NewNopLogger())
	dk := keeper.NewKeeper(encCfg.Codec, runtime.NewKVStoreService(keys[types.StoreKey]), ak, bk, nil, "fee_collector", authtypes.NewModuleAddress("gov").String())
	commission := make(sdk.DecCoins, n)
	for i := range commission {
		denom := fmt.Sprintf("token%03d", i)
		commission[i] = sdk.NewDecCoinFromDec(denom, math.LegacyNewDecWithPrec(125, 2))
		require.NoError(tb, bk.Balances.Set(ctx, collections.Join(module.GetAddress(), denom), math.NewInt(2)))
	}
	valAddr := sdk.ValAddress(addr)
	require.NoError(tb, dk.SetValidatorAccumulatedCommission(ctx, valAddr, types.ValidatorAccumulatedCommission{Commission: commission}))
	require.NoError(tb, dk.SetValidatorOutstandingRewards(ctx, valAddr, types.ValidatorOutstandingRewards{Rewards: commission}))
	cms.Commit()
	return ctx, dk, bk, valAddr
}

func TestWithdrawValidatorCommissionManyDenoms(t *testing.T) {
	for _, n := range []int{1, 10, 100} {
		t.Run(fmt.Sprintf("N=%d", n), func(t *testing.T) {
			parent, dk, bk, addr := setupCommissionWithdrawal(t, n)
			ctx, _ := parent.CacheContext()
			ctx = ctx.WithGasMeter(storetypes.NewInfiniteGasMeter())
			coins, err := dk.WithdrawValidatorCommission(ctx, addr)
			require.NoError(t, err)
			t.Logf("withdrawal gas: %d", ctx.GasMeter().GasConsumed())
			wantCoins := make(sdk.Coins, n)
			wantRemainder := make(sdk.DecCoins, n)
			for i := range wantCoins {
				denom := fmt.Sprintf("token%03d", i)
				wantCoins[i] = sdk.NewInt64Coin(denom, 1)
				wantRemainder[i] = sdk.NewDecCoinFromDec(denom, math.LegacyNewDecWithPrec(25, 2))
			}
			require.Equal(t, wantCoins, coins)
			require.Equal(t, wantCoins, bk.GetAllBalances(ctx, sdk.AccAddress(addr)))
			require.Equal(t, wantCoins, bk.GetAllBalances(ctx, authtypes.NewModuleAddress(types.ModuleName)))
			commission, err := dk.GetValidatorAccumulatedCommission(ctx, addr)
			require.NoError(t, err)
			require.Equal(t, wantRemainder, commission.Commission)
			outstanding, err := dk.GetValidatorOutstandingRewards(ctx, addr)
			require.NoError(t, err)
			require.Equal(t, wantRemainder, outstanding.Rewards)
			events := ctx.EventManager().Events()
			require.Contains(t, events, sdk.NewEvent(types.EventTypeWithdrawCommission, sdk.NewAttribute(sdk.AttributeKeyAmount, wantCoins.String())))
			// The fractional remainder cannot be paid out a second time.
			again, err := dk.WithdrawValidatorCommission(ctx, addr)
			require.NoError(t, err)
			require.Nil(t, again)
			require.Equal(t, wantCoins, bk.GetAllBalances(ctx, sdk.AccAddress(addr)))
		})
	}
}

func BenchmarkWithdrawValidatorCommission(b *testing.B) {
	for _, n := range []int{1, 10, 100} {
		b.Run(fmt.Sprintf("N=%d", n), func(b *testing.B) {
			parent, dk, _, addr := setupCommissionWithdrawal(b, n)
			b.ReportAllocs()
			for b.Loop() {
				// Discard each transaction's writes so every iteration withdraws
				// the same amounts from the same initial state.
				ctx, _ := parent.CacheContext()
				if _, err := dk.WithdrawValidatorCommission(ctx, addr); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
