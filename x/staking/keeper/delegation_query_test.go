package keeper_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtestutil "github.com/cosmos/cosmos-sdk/x/staking/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

type delegationQueryStoreService struct {
	corestore.KVStoreService
	validatorReads int
	paramsReads    int
}

func (s *delegationQueryStoreService) OpenKVStore(ctx context.Context) corestore.KVStore {
	return delegationQueryStore{s.KVStoreService.OpenKVStore(ctx), s}
}

type delegationQueryStore struct {
	corestore.KVStore
	service *delegationQueryStoreService
}

func (s delegationQueryStore) Get(key []byte) ([]byte, error) {
	if bytes.HasPrefix(key, types.ValidatorsKey) {
		s.service.validatorReads++
	}
	if bytes.Equal(key, types.ParamsKey) {
		s.service.paramsReads++
	}
	return s.KVStore.Get(key)
}

type delegationQueryFixture struct {
	ctx        sdk.Context
	keeper     *stakingkeeper.Keeper
	store      *delegationQueryStoreService
	validators []types.Validator
	delegator  string
}

func newDelegationQueryFixture(tb testing.TB, n int, distinctValidators bool) delegationQueryFixture {
	tb.Helper()
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(tb, key, storetypes.NewTransientStoreKey("transient_test"))
	tb.Cleanup(func() { require.NoError(tb, testCtx.DB.Close()) })
	encCfg := moduletestutil.MakeTestEncodingConfig()
	types.RegisterInterfaces(encCfg.InterfaceRegistry)
	service := &delegationQueryStoreService{KVStoreService: runtime.NewKVStoreService(key)}
	accountKeeper := stakingtestutil.NewMockAccountKeeper(gomock.NewController(tb))
	accountKeeper.EXPECT().GetModuleAddress(gomock.Any()).DoAndReturn(authtypes.NewModuleAddress).AnyTimes()
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()
	k := stakingkeeper.NewKeeper(encCfg.Codec, service, accountKeeper, nil,
		authtypes.NewModuleAddress("gov").String(), address.NewBech32Codec("cosmosvaloper"), address.NewBech32Codec("cosmosvalcons"))
	require.NoError(tb, k.SetParams(testCtx.Ctx, types.DefaultParams()))
	f := delegationQueryFixture{ctx: testCtx.Ctx, keeper: k, store: service}
	for i := range n {
		valIndex, delIndex := 0, i
		if distinctValidators {
			valIndex, delIndex = i, 0
		}
		valBytes, delBytes := make([]byte, 20), make([]byte, 20)
		binary.BigEndian.PutUint64(valBytes[12:], uint64(valIndex+1))
		binary.BigEndian.PutUint64(delBytes[12:], uint64(delIndex+1))
		valAddr, delAddr := sdk.ValAddress(valBytes).String(), sdk.AccAddress(delBytes).String()
		if i == 0 || distinctValidators {
			v, err := types.NewValidator(valAddr, PKs[valIndex], types.Description{})
			require.NoError(tb, err)
			v.Tokens = math.NewInt(1_000_000)
			v.DelegatorShares = math.LegacyNewDec(2_000_000)
			require.NoError(tb, k.SetValidator(f.ctx, v))
			f.validators = append(f.validators, v)
		}
		require.NoError(tb, k.SetDelegation(f.ctx, types.NewDelegation(delAddr, valAddr, math.LegacyNewDec(int64(2*i+3)))))
		f.delegator = delAddr
	}
	f.store.validatorReads, f.store.paramsReads = 0, 0
	f.ctx = f.ctx.WithGasMeter(storetypes.NewInfiniteGasMeter())
	return f
}

func (f delegationQueryFixture) query(distinctValidators bool, page *query.PageRequest) ([]types.DelegationResponse, *query.PageResponse, error) {
	q := stakingkeeper.Querier{Keeper: f.keeper}
	if distinctValidators {
		res, err := q.DelegatorDelegations(f.ctx, &types.QueryDelegatorDelegationsRequest{DelegatorAddr: f.delegator, Pagination: page})
		if err != nil {
			return nil, nil, err
		}
		return res.DelegationResponses, res.Pagination, nil
	}
	res, err := q.ValidatorDelegations(f.ctx, &types.QueryValidatorDelegationsRequest{ValidatorAddr: f.validators[0].OperatorAddress, Pagination: page})
	if err != nil {
		return nil, nil, err
	}
	return res.DelegationResponses, res.Pagination, nil
}

func TestDelegationQueriesReuseRecords(t *testing.T) {
	for _, distinct := range []bool{false, true} {
		for _, n := range []int{1, 10, 100} {
			t.Run(fmt.Sprintf("distinct=%t/N=%d", distinct, n), func(t *testing.T) {
				f := newDelegationQueryFixture(t, n, distinct)
				responses, page, err := f.query(distinct, &query.PageRequest{Limit: uint64(n), CountTotal: true})
				require.NoError(t, err)
				require.Len(t, responses, n)
				require.Equal(t, uint64(n), page.Total)
				require.Empty(t, page.NextKey)
				for i, response := range responses {
					require.Equal(t, math.LegacyNewDec(int64(2*i+3)), response.Delegation.Shares)
					require.Equal(t, sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(i+1)), response.Balance)
				}
				require.Equal(t, len(f.validators), f.store.validatorReads)
				require.Equal(t, 1, f.store.paramsReads)
			})
		}
	}
}

func TestDelegationQueriesPaginationAndUpdates(t *testing.T) {
	for _, distinct := range []bool{false, true} {
		t.Run(fmt.Sprintf("distinct=%t", distinct), func(t *testing.T) {
			f := newDelegationQueryFixture(t, 10, distinct)
			first, page, err := f.query(distinct, &query.PageRequest{Limit: 3, CountTotal: true})
			require.NoError(t, err)
			require.Len(t, first, 3)
			require.Equal(t, uint64(10), page.Total)
			require.NotEmpty(t, page.NextKey)
			next, _, err := f.query(distinct, &query.PageRequest{Key: page.NextKey, Limit: 3})
			require.NoError(t, err)
			require.Len(t, next, 3)
			require.Equal(t, int64(4), next[0].Balance.Amount.Int64())
			reverse, _, err := f.query(distinct, &query.PageRequest{Limit: 3, Reverse: true})
			require.NoError(t, err)
			require.Equal(t, int64(10), reverse[0].Balance.Amount.Int64())

			for _, validator := range f.validators {
				validator.Tokens = math.NewInt(2_000_000)
				require.NoError(t, f.keeper.SetValidator(f.ctx, validator))
			}
			params := types.DefaultParams()
			params.BondDenom = "updated"
			require.NoError(t, f.keeper.SetParams(f.ctx, params))
			updated, _, err := f.query(distinct, &query.PageRequest{Limit: 10})
			require.NoError(t, err)
			for i, response := range updated {
				require.Equal(t, sdk.NewInt64Coin("updated", int64(2*i+3)), response.Balance)
			}
		})
	}
}

func TestDelegationQueriesEmptyPageAndMissingValidator(t *testing.T) {
	for _, distinct := range []bool{false, true} {
		t.Run(fmt.Sprintf("distinct=%t", distinct), func(t *testing.T) {
			f := newDelegationQueryFixture(t, 1, distinct)
			valAddr, err := sdk.ValAddressFromBech32(f.validators[0].OperatorAddress)
			require.NoError(t, err)
			require.NoError(t, f.store.OpenKVStore(f.ctx).Delete(types.GetValidatorKey(valAddr)))

			// An empty page must not try to load a validator or module parameters.
			responses, _, err := f.query(distinct, &query.PageRequest{Offset: 1, Limit: 1})
			require.NoError(t, err)
			require.Empty(t, responses)
			require.NotNil(t, responses)
			require.Zero(t, f.store.validatorReads)
			require.Zero(t, f.store.paramsReads)

			responses, _, err = f.query(distinct, &query.PageRequest{Limit: 1})
			require.Equal(t, codes.Internal, status.Code(err))
			require.Nil(t, responses)
			require.ErrorContains(t, err, types.ErrNoValidatorFound.Error())
			require.Equal(t, 1, f.store.validatorReads)
			require.Zero(t, f.store.paramsReads)
		})
	}
}

func BenchmarkDelegationQueries(b *testing.B) {
	for _, distinct := range []bool{false, true} {
		for _, n := range []int{1, 10, 100} {
			b.Run(fmt.Sprintf("distinct=%t/N=%d", distinct, n), func(b *testing.B) {
				f := newDelegationQueryFixture(b, n, distinct)
				page := &query.PageRequest{Limit: uint64(n)}
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					responses, _, err := f.query(distinct, page)
					if err != nil || len(responses) != n {
						b.Fatalf("query returned %d responses: %v", len(responses), err)
					}
				}
				b.StopTimer()
				b.ReportMetric(float64(f.store.validatorReads)/float64(b.N), "validator-reads/op")
				b.ReportMetric(float64(f.store.paramsReads)/float64(b.N), "params-reads/op")
				b.ReportMetric(float64(f.ctx.GasMeter().GasConsumed())/float64(b.N), "gas/op")
			})
		}
	}
}
