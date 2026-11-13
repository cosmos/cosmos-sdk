package keeper_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	coreaddress "cosmossdk.io/core/address"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/runtime"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtestutil "github.com/cosmos/cosmos-sdk/x/distribution/testutil"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type rewardAggregationAccount struct{ disttypes.AccountKeeper }

func (rewardAggregationAccount) AddressCodec() coreaddress.Codec {
	return address.NewBech32Codec(sdk.Bech32MainPrefix)
}

func (rewardAggregationAccount) GetModuleAddress(name string) sdk.AccAddress {
	return authtypes.NewModuleAddress(name)
}

// Keep staking lookups deterministic while exercising the real distribution
// keeper and store in the query benchmarks.
type rewardAggregationStaking struct {
	disttypes.StakingKeeper
	validators  map[string]stakingtypes.Validator
	delegations []stakingtypes.Delegation
}

func (*rewardAggregationStaking) ValidatorAddressCodec() coreaddress.Codec {
	return address.NewBech32Codec(sdk.Bech32PrefixValAddr)
}

func (s *rewardAggregationStaking) Validator(_ context.Context, addr sdk.ValAddress) (stakingtypes.ValidatorI, error) {
	v := s.validators[string(addr)]
	return v, nil
}

func (s *rewardAggregationStaking) Delegation(_ context.Context, del sdk.AccAddress, val sdk.ValAddress) (stakingtypes.DelegationI, error) {
	v := s.validators[string(val)]
	return stakingtypes.NewDelegation(del.String(), val.String(), v.DelegatorShares), nil
}

func (s *rewardAggregationStaking) IterateDelegations(_ context.Context, _ sdk.AccAddress, fn func(int64, stakingtypes.DelegationI) bool) error {
	for i, d := range s.delegations {
		if fn(int64(i), d) {
			break
		}
	}
	return nil
}

type rewardAggregationFixture struct {
	ctx           sdk.Context
	q             keeper.Querier
	req           *disttypes.QueryDelegationTotalRewardsRequest
	staking       *rewardAggregationStaking
	wantTotal     sdk.DecCoins
	wantClaimable sdk.Coins
	wantRewards   []sdk.DecCoins
}

func setupRewardAggregation(tb testing.TB, n, width int, disjoint bool, amounts ...string) *rewardAggregationFixture {
	tb.Helper()
	key := storetypes.NewKVStoreKey(disttypes.StoreKey)
	tc := testutil.DefaultContextWithDB(tb, key, storetypes.NewTransientStoreKey("reward_aggregation"))
	tb.Cleanup(func() { require.NoError(tb, tc.DB.Close()) })
	ctx := tc.Ctx.WithBlockHeight(1)
	service := runtime.NewKVStoreService(key)
	enc := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})
	s := &rewardAggregationStaking{validators: make(map[string]stakingtypes.Validator)}
	k := keeper.NewKeeper(enc.Codec, service, rewardAggregationAccount{}, nil, s, "fee_collector", authtypes.NewModuleAddress("gov").String())
	require.NoError(tb, k.FeePool.Set(ctx, disttypes.InitialFeePool()))
	require.NoError(tb, k.Params.Set(ctx, disttypes.DefaultParams()))
	del := sdk.AccAddress(ed25519.GenPrivKeyFromSecret([]byte("reward-aggregation-delegator")).PubKey().Address())
	f := &rewardAggregationFixture{
		wantTotal: sdk.DecCoins{}, wantClaimable: sdk.Coins{},
		q:       keeper.NewQuerier(k),
		req:     &disttypes.QueryDelegationTotalRewardsRequest{DelegatorAddress: del.String()},
		staking: s,
	}
	for i := 0; i < n; i++ {
		pk := ed25519.GenPrivKeyFromSecret([]byte(fmt.Sprintf("reward-aggregation-validator-%d", i))).PubKey()
		v, err := distrtestutil.CreateValidator(pk, math.NewInt(1000))
		require.NoError(tb, err)
		v.Commission = stakingtypes.NewCommission(math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec())
		val := sdk.ValAddress(sdk.GetConsAddress(pk))
		s.validators[string(val)] = v
		s.delegations = append(s.delegations, stakingtypes.NewDelegation(del.String(), val.String(), v.DelegatorShares))
		require.NoError(tb, distrtestutil.CallCreateValidatorHooks(ctx, k, del, val))
		amount := "1.6"
		if len(amounts) > 0 {
			amount = amounts[i%len(amounts)]
		}
		rewardAmount := math.LegacyMustNewDecFromStr(amount)
		count := width
		if rewardAmount.IsZero() {
			count = 0
		}
		coins := make(sdk.DecCoins, count)
		for j := range coins {
			index := j
			if disjoint {
				index += i * width
			}
			coins[j] = sdk.DecCoin{Denom: fmt.Sprintf("token%04d", index), Amount: rewardAmount}
		}
		f.wantRewards = append(f.wantRewards, coins)
		f.wantTotal = f.wantTotal.Add(coins...)
		truncated, _ := coins.TruncateDecimal()
		f.wantClaimable = f.wantClaimable.Add(truncated...)
		require.NoError(tb, k.AllocateTokensToValidator(ctx.WithBlockHeight(2), v, coins))
	}
	f.ctx = ctx.WithBlockHeight(2)
	return f
}

func (f *rewardAggregationFixture) query() (*disttypes.QueryDelegationTotalRewardsResponse, error) {
	ctx, _ := f.ctx.CacheContext()
	return f.q.DelegationTotalRewards(ctx, f.req)
}

func (f *rewardAggregationFixture) check(tb testing.TB, response *disttypes.QueryDelegationTotalRewardsResponse) {
	tb.Helper()
	require.Equal(tb, f.wantTotal, response.Total, "raw total or empty response shape mismatch")
	require.True(tb, f.wantClaimable.Equal(response.Claimable), "claimable total mismatch")
	require.Len(tb, response.Rewards, len(f.wantRewards))
	for i, r := range response.Rewards {
		require.True(tb, f.wantRewards[i].Equal(r.Reward), "per-delegation reward mismatch")
		require.Equal(tb, f.staking.delegations[i].ValidatorAddress, r.ValidatorAddress)
	}
}

func TestQueryDelegationTotalRewardsAggregation(t *testing.T) {
	for _, disjoint := range []bool{false, true} {
		for _, width := range []int{1, 10} {
			for _, n := range []int{1, 10, 100} {
				name := fmt.Sprintf("disjoint=%t/denoms=%d/N=%d", disjoint, width, n)
				t.Run(name, func(t *testing.T) {
					f := setupRewardAggregation(t, n, width, disjoint)
					response, err := f.query()
					require.NoError(t, err)
					f.check(t, response)
				})
			}
		}
	}
}

func BenchmarkQueryDelegationTotalRewards(b *testing.B) {
	for _, disjoint := range []bool{false, true} {
		for _, width := range []int{1, 10} {
			for _, n := range []int{1, 10, 100} {
				b.Run(fmt.Sprintf("disjoint=%t/denoms=%d/N=%d", disjoint, width, n), func(b *testing.B) {
					f := setupRewardAggregation(b, n, width, disjoint)
					response, err := f.query()
					require.NoError(b, err)
					f.check(b, response)
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						if _, err := f.query(); err != nil {
							b.Fatal(err)
						}
					}
					b.StopTimer()
					b.ReportMetric(float64(n*width), "reward_terms/op")
				})
			}
		}
	}
}

func TestQueryDelegationTotalRewardsBoundaries(t *testing.T) {
	for _, amounts := range [][]string{{"0"}, {"0.6"}, {"1"}, {"1.999"}, {"0", "0.6", "1.999"}, {"123456789123456789.999"}} {
		for _, n := range []int{0, 1, 10} {
			for _, disjoint := range []bool{false, true} {
				t.Run(fmt.Sprintf("amounts=%v/N=%d/disjoint=%t", amounts, n, disjoint), func(t *testing.T) {
					f := setupRewardAggregation(t, n, 3, disjoint, amounts...)
					response, err := f.query()
					require.NoError(t, err)
					f.check(t, response)
					// Queries use a cache context and must not persist period changes.
					again, err := f.query()
					require.NoError(t, err)
					require.Equal(t, response, again)
				})
			}
		}
	}
}
