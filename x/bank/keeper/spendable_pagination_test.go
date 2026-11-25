package keeper_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/collections"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/log/v2"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/store/v2"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/types/query"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// Count values read through the bank store, including reads outside the requested page.
type balanceReadService struct {
	corestore.KVStoreService
	values int
}

func (s *balanceReadService) OpenKVStore(ctx context.Context) corestore.KVStore {
	return balanceReadStore{s.KVStoreService.OpenKVStore(ctx), &s.values}
}

type balanceReadStore struct {
	corestore.KVStore
	values *int
}

func (s balanceReadStore) Iterator(start, end []byte) (corestore.Iterator, error) {
	it, err := s.KVStore.Iterator(start, end)
	if err != nil {
		return nil, err
	}
	return balanceReadIterator{it, s.values}, nil
}

func (s balanceReadStore) ReverseIterator(start, end []byte) (corestore.Iterator, error) {
	it, err := s.KVStore.ReverseIterator(start, end)
	if err != nil {
		return nil, err
	}
	return balanceReadIterator{it, s.values}, nil
}

type balanceReadIterator struct {
	corestore.Iterator
	values *int
}

func (it balanceReadIterator) Value() []byte {
	*it.values++
	return it.Iterator.Value()
}

func setupSpendablePagination(tb testing.TB, n int, locked sdk.Coins) (sdk.Context, keeper.BaseKeeper, banktypes.QueryClient, sdk.AccAddress, *balanceReadService) {
	tb.Helper()
	key := storetypes.NewKVStoreKey(banktypes.StoreKey)
	authKey := storetypes.NewKVStoreKey(authtypes.StoreKey)
	db := dbm.NewMemDB()
	tb.Cleanup(func() { require.NoError(tb, db.Close()) })
	cms := store.NewCommitMultiStore(db, log.NewNopLogger())
	cms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, nil)
	cms.MountStoreWithDB(authKey, storetypes.StoreTypeIAVL, nil)
	require.NoError(tb, cms.LoadLatestVersion())
	ctx := sdk.NewContext(cms, cmtproto.Header{Time: time.Unix(1, 0)}, false, log.NewNopLogger())
	addr := sdk.AccAddress([]byte("pagination-account__"))
	var acc sdk.AccountI = authtypes.NewBaseAccountWithAddress(addr)
	if len(locked) > 0 {
		var err error
		acc, err = vestingtypes.NewDelayedVestingAccount(acc.(*authtypes.BaseAccount), locked, 100)
		require.NoError(tb, err)
	}
	encCfg := moduletestutil.MakeTestEncodingConfig()
	authtypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	vestingtypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	ak := authkeeper.NewAccountKeeper(encCfg.Codec, runtime.NewKVStoreService(authKey), authtypes.ProtoBaseAccount, nil, address.NewBech32Codec("cosmos"), "cosmos", authtypes.NewModuleAddress("gov").String())
	ak.SetAccount(ctx, acc)
	service := &balanceReadService{KVStoreService: runtime.NewKVStoreService(key)}
	k := keeper.NewBaseKeeper(encCfg.Codec, service, ak, nil, authtypes.NewModuleAddress("gov").String(), log.NewNopLogger())
	for i := range n {
		require.NoError(tb, k.Balances.Set(ctx, collections.Join(addr, fmt.Sprintf("token%03d", i)), math.NewInt(100)))
	}
	cms.Commit()
	banktypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	helper := baseapp.NewQueryServerTestHelper(ctx, encCfg.InterfaceRegistry)
	banktypes.RegisterQueryServer(helper, k)
	service.values = 0
	return ctx, k, banktypes.NewQueryClient(helper), addr, service
}

func TestSpendableBalancesPageReads(t *testing.T) {
	for _, n := range []int{1, 10, 100} {
		t.Run(fmt.Sprintf("N=%d", n), func(t *testing.T) {
			ctx, _, client, addr, reads := setupSpendablePagination(t, n, nil)
			res, err := client.SpendableBalances(ctx, &banktypes.QuerySpendableBalancesRequest{
				Address: addr.String(), Pagination: &query.PageRequest{Limit: 1},
			})
			require.NoError(t, err)
			require.Equal(t, sdk.Coins{sdk.NewInt64Coin("token000", 100)}, res.Balances)
			require.Equal(t, n > 1, len(res.Pagination.NextKey) > 0)
			require.Equal(t, 1, reads.values, "a one-balance page must not decode every other balance")
		})
	}
}

func TestSpendableBalancesPaginationErrors(t *testing.T) {
	ctx, k, _, addr, reads := setupSpendablePagination(t, 3, nil)
	for _, req := range []*banktypes.QuerySpendableBalancesRequest{
		nil,
		{Address: "invalid-address"},
		{Address: addr.String(), Pagination: &query.PageRequest{Key: []byte("key"), Offset: 1}},
	} {
		_, err := k.SpendableBalances(ctx, req)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	}
	require.Zero(t, reads.values)

	// Balances can be queried even when there is no corresponding auth account.
	missing := sdk.AccAddress([]byte("missing-account_____"))
	require.NoError(t, k.Balances.Set(ctx, collections.Join(missing, "token000"), math.NewInt(100)))
	res, err := k.SpendableBalances(ctx, &banktypes.QuerySpendableBalancesRequest{Address: missing.String()})
	require.NoError(t, err)
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin("token000", 100)}, res.Balances)
}

func TestSpendableBalancesPaginationCompatibility(t *testing.T) {
	for _, tc := range []struct {
		name   string
		locked sdk.Coins
		want   []int64
	}{
		{"unlocked", nil, []int64{100, 100, 100}},
		{"partially locked", sdk.NewCoins(sdk.NewInt64Coin("token001", 25)), []int64{100, 75, 100}},
		{"fully locked denom", sdk.NewCoins(sdk.NewInt64Coin("token001", 100)), []int64{100, 0, 100}},
		// Preserve the current all-zero response when any locked amount exceeds its balance,
		// even if that denomination is outside the requested page or absent from the account.
		{"off-page deficit", sdk.NewCoins(sdk.NewInt64Coin("token002", 101)), []int64{0, 0, 0}},
		{"absent locked denom", sdk.NewCoins(sdk.NewInt64Coin("token999", 1)), []int64{0, 0, 0}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, k, client, addr, _ := setupSpendablePagination(t, 3, tc.locked)
			for _, reverse := range []bool{false, true} {
				var next []byte
				for page := range 3 {
					res, err := client.SpendableBalances(ctx, &banktypes.QuerySpendableBalancesRequest{
						Address: addr.String(), Pagination: &query.PageRequest{Limit: 1, Key: next, Reverse: reverse},
					})
					require.NoError(t, err)
					i := page
					if reverse {
						i = 2 - page
					}
					require.Equal(t, sdk.Coins{sdk.NewInt64Coin(fmt.Sprintf("token%03d", i), tc.want[i])}, res.Balances)
					next = res.Pagination.NextKey
					require.Equal(t, page < 2, len(next) > 0)
				}
			}
			res, err := client.SpendableBalances(ctx, &banktypes.QuerySpendableBalancesRequest{
				Address: addr.String(), Pagination: &query.PageRequest{Offset: 1, Limit: 1, CountTotal: true},
			})
			require.NoError(t, err)
			require.Equal(t, sdk.Coins{sdk.NewInt64Coin("token001", tc.want[1])}, res.Balances)
			require.Equal(t, uint64(3), res.Pagination.Total)
			// Wrapped contexts must also expose full balances once vesting ends.
			wrapped := context.WithValue(context.Background(), sdk.SdkContextKey, ctx.WithBlockTime(time.Unix(101, 0)))
			res, err = k.SpendableBalances(wrapped, &banktypes.QuerySpendableBalancesRequest{Address: addr.String()})
			require.NoError(t, err)
			require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("token000", 100), sdk.NewInt64Coin("token001", 100), sdk.NewInt64Coin("token002", 100)), res.Balances)
		})
	}
	ctx, k, _, addr, _ := setupSpendablePagination(t, 0, nil)
	res, err := k.SpendableBalances(ctx, &banktypes.QuerySpendableBalancesRequest{Address: addr.String()})
	require.NoError(t, err)
	require.NotNil(t, res.Balances)
	require.Empty(t, res.Balances)
}

func BenchmarkSpendableBalancesPagination(b *testing.B) {
	for _, n := range []int{1, 10, 100} {
		for _, vesting := range []bool{false, true} {
			for _, mode := range []string{"handler", "query-router", "all-pages"} {
				b.Run(fmt.Sprintf("N=%d/vesting=%t/%s", n, vesting, mode), func(b *testing.B) {
					var locked sdk.Coins
					if vesting {
						locked = sdk.NewCoins(sdk.NewInt64Coin(fmt.Sprintf("token%03d", n-1), 25))
					}
					ctx, k, client, addr, reads := setupSpendablePagination(b, n, locked)
					req := &banktypes.QuerySpendableBalancesRequest{Address: addr.String(), Pagination: &query.PageRequest{Limit: 1}}
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						req.Pagination.Key = nil
						for {
							var res *banktypes.QuerySpendableBalancesResponse
							var err error
							if mode == "handler" {
								res, err = k.SpendableBalances(ctx, req)
							} else {
								res, err = client.SpendableBalances(ctx, req)
							}
							if err != nil || len(res.Balances) != 1 {
								b.Fatalf("unexpected response: %v, %v", res, err)
							}
							if mode != "all-pages" || len(res.Pagination.NextKey) == 0 {
								break
							}
							req.Pagination.Key = res.Pagination.NextKey
						}
					}
					b.StopTimer()
					b.ReportMetric(float64(reads.values)/float64(b.N), "balance-values/op")
				})
			}
		}
	}
}
