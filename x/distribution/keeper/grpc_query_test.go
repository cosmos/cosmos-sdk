package keeper_test

import (
	"testing"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
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

type validatorQueryTestFixture struct {
	ctx         sdk.Context
	distrKeeper keeper.Keeper
	querier     keeper.Querier
	valAddr     sdk.ValAddress
	addr        sdk.AccAddress
	val         stakingtypes.Validator
}

func setupValidatorQueryTest(t *testing.T, expectNonExistentValidator bool) *validatorQueryTestFixture {
	t.Helper()
	ctrl := gomock.NewController(t)
	key := storetypes.NewKVStoreKey(disttypes.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Height: 1})

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())
	stakingKeeper.EXPECT().ValidatorAddressCodec().Return(address.NewBech32Codec(sdk.Bech32PrefixValAddr)).AnyTimes()
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec(sdk.Bech32MainPrefix)).AnyTimes()

	distrKeeper := keeper.NewKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		"fee_collector",
		authtypes.NewModuleAddress("gov").String(),
	)

	require.NoError(t, distrKeeper.FeePool.Set(ctx, disttypes.InitialFeePool()))
	require.NoError(t, distrKeeper.Params.Set(ctx, disttypes.DefaultParams()))

	valAddr := sdk.ValAddress(valConsAddr0)
	addr := sdk.AccAddress(valAddr)
	val, err := distrtestutil.CreateValidator(valConsPk0, math.NewInt(1000))
	require.NoError(t, err)
	val.Commission = stakingtypes.NewCommission(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0))

	del := stakingtypes.NewDelegation(addr.String(), valAddr.String(), val.DelegatorShares)

	stakingKeeper.EXPECT().Validator(gomock.Any(), valAddr).Return(val, nil).AnyTimes()
	if expectNonExistentValidator {
		stakingKeeper.EXPECT().Validator(gomock.Any(), gomock.Not(gomock.Eq(valAddr))).Return(nil, nil).AnyTimes()
	}
	stakingKeeper.EXPECT().Delegation(gomock.Any(), addr, valAddr).Return(del, nil).AnyTimes()

	err = distrtestutil.CallCreateValidatorHooks(ctx, distrKeeper, addr, valAddr)
	require.NoError(t, err)

	return &validatorQueryTestFixture{
		ctx:         ctx,
		distrKeeper: distrKeeper,
		querier:     keeper.NewQuerier(distrKeeper),
		valAddr:     valAddr,
		addr:        addr,
		val:         val,
	}
}

func (f *validatorQueryTestFixture) allocateRewardsAndIncrementPeriod(t *testing.T, amount int64) {
	t.Helper()
	tokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(amount)}}
	require.NoError(t, f.distrKeeper.AllocateTokensToValidator(f.ctx, f.val, tokens))
	f.ctx = f.ctx.WithBlockHeight(f.ctx.BlockHeight() + 1)
	_, err := f.distrKeeper.IncrementValidatorPeriod(f.ctx, f.val)
	require.NoError(t, err)
}

func TestQueryValidatorHistoricalRewards(t *testing.T) {
	f := setupValidatorQueryTest(t, false)
	f.allocateRewardsAndIncrementPeriod(t, 100)
	f.allocateRewardsAndIncrementPeriod(t, 200)

	testCases := []struct {
		name               string
		req                *disttypes.QueryValidatorHistoricalRewardsRequest
		expectErr          bool
		errContains        string
		expectedRefCount   uint32
		expectNonZeroRatio bool
	}{
		{
			name: "period 1 has reference count 1",
			req: &disttypes.QueryValidatorHistoricalRewardsRequest{
				ValidatorAddress: f.valAddr.String(),
				Period:           1,
			},
			expectErr:          false,
			expectedRefCount:   1,
			expectNonZeroRatio: false,
		},
		{
			name: "period 2 deleted (reference count 0)",
			req: &disttypes.QueryValidatorHistoricalRewardsRequest{
				ValidatorAddress: f.valAddr.String(),
				Period:           2,
			},
			expectErr:          false,
			expectedRefCount:   0,
			expectNonZeroRatio: false,
		},
		{
			name: "period 3 is current with rewards",
			req: &disttypes.QueryValidatorHistoricalRewardsRequest{
				ValidatorAddress: f.valAddr.String(),
				Period:           3,
			},
			expectErr:          false,
			expectedRefCount:   1,
			expectNonZeroRatio: true,
		},
		{
			name:        "nil request",
			req:         nil,
			expectErr:   true,
			errContains: "invalid request",
		},
		{
			name: "empty validator address",
			req: &disttypes.QueryValidatorHistoricalRewardsRequest{
				ValidatorAddress: "",
				Period:           0,
			},
			expectErr:   true,
			errContains: "empty validator address",
		},
		{
			name: "invalid validator address",
			req: &disttypes.QueryValidatorHistoricalRewardsRequest{
				ValidatorAddress: "invalid",
				Period:           0,
			},
			expectErr: true,
		},
		{
			name: "non-existent period returns empty",
			req: &disttypes.QueryValidatorHistoricalRewardsRequest{
				ValidatorAddress: f.valAddr.String(),
				Period:           999,
			},
			expectErr:          false,
			expectedRefCount:   0,
			expectNonZeroRatio: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := f.querier.ValidatorHistoricalRewards(f.ctx, tc.req)
			if tc.expectErr {
				require.Error(t, err)
				if tc.errContains != "" {
					require.Contains(t, err.Error(), tc.errContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.Equal(t, tc.expectedRefCount, resp.Rewards.ReferenceCount)
				if tc.expectNonZeroRatio {
					require.False(t, resp.Rewards.CumulativeRewardRatio.IsZero())
				} else {
					require.True(t, resp.Rewards.CumulativeRewardRatio.IsZero())
				}
			}
		})
	}
}

func TestQueryValidatorCurrentRewards(t *testing.T) {
	f := setupValidatorQueryTest(t, true)
	f.allocateRewardsAndIncrementPeriod(t, 100)

	testCases := []struct {
		name                 string
		req                  *disttypes.QueryValidatorCurrentRewardsRequest
		expectErr            bool
		errContains          string
		expectedPeriod       uint64
		expectNonZeroRewards bool
	}{
		{
			name: "valid validator with current rewards",
			req: &disttypes.QueryValidatorCurrentRewardsRequest{
				ValidatorAddress: f.valAddr.String(),
			},
			expectErr:            false,
			expectedPeriod:       3,
			expectNonZeroRewards: false,
		},
		{
			name:        "nil request",
			req:         nil,
			expectErr:   true,
			errContains: "invalid request",
		},
		{
			name: "empty validator address",
			req: &disttypes.QueryValidatorCurrentRewardsRequest{
				ValidatorAddress: "",
			},
			expectErr:   true,
			errContains: "empty validator address",
		},
		{
			name: "invalid validator address",
			req: &disttypes.QueryValidatorCurrentRewardsRequest{
				ValidatorAddress: "invalid",
			},
			expectErr: true,
		},
		{
			name: "non-existent validator",
			req: &disttypes.QueryValidatorCurrentRewardsRequest{
				ValidatorAddress: sdk.ValAddress([]byte("nonexistent")).String(),
			},
			expectErr:   true,
			errContains: "validator does not exist",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := f.querier.ValidatorCurrentRewards(f.ctx, tc.req)
			if tc.expectErr {
				require.Error(t, err)
				if tc.errContains != "" {
					require.Contains(t, err.Error(), tc.errContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.Equal(t, tc.expectedPeriod, resp.Rewards.Period)
				if tc.expectNonZeroRewards {
					require.False(t, resp.Rewards.Rewards.IsZero())
				} else {
					require.True(t, resp.Rewards.Rewards.IsZero())
				}
			}
		})
	}
}

func TestQueryDelegatorStartingInfo(t *testing.T) {
	f := setupValidatorQueryTest(t, true)
	f.allocateRewardsAndIncrementPeriod(t, 100)

	testCases := []struct {
		name           string
		req            *disttypes.QueryDelegatorStartingInfoRequest
		expectErr      bool
		errContains    string
		expectedPeriod uint64
		expectedHeight uint64
	}{
		{
			name: "valid delegator and validator",
			req: &disttypes.QueryDelegatorStartingInfoRequest{
				DelegatorAddress: f.addr.String(),
				ValidatorAddress: f.valAddr.String(),
			},
			expectErr:      false,
			expectedPeriod: 1,
			expectedHeight: 1,
		},
		{
			name:        "nil request",
			req:         nil,
			expectErr:   true,
			errContains: "invalid request",
		},
		{
			name: "empty delegator address",
			req: &disttypes.QueryDelegatorStartingInfoRequest{
				DelegatorAddress: "",
				ValidatorAddress: f.valAddr.String(),
			},
			expectErr:   true,
			errContains: "empty delegator address",
		},
		{
			name: "empty validator address",
			req: &disttypes.QueryDelegatorStartingInfoRequest{
				DelegatorAddress: f.addr.String(),
				ValidatorAddress: "",
			},
			expectErr:   true,
			errContains: "empty validator address",
		},
		{
			name: "invalid delegator address",
			req: &disttypes.QueryDelegatorStartingInfoRequest{
				DelegatorAddress: "invalid",
				ValidatorAddress: f.valAddr.String(),
			},
			expectErr: true,
		},
		{
			name: "invalid validator address",
			req: &disttypes.QueryDelegatorStartingInfoRequest{
				DelegatorAddress: f.addr.String(),
				ValidatorAddress: "invalid",
			},
			expectErr: true,
		},
		{
			name: "non-existent validator",
			req: &disttypes.QueryDelegatorStartingInfoRequest{
				DelegatorAddress: f.addr.String(),
				ValidatorAddress: sdk.ValAddress([]byte("nonexistent")).String(),
			},
			expectErr:   true,
			errContains: "validator does not exist",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := f.querier.DelegatorStartingInfo(f.ctx, tc.req)
			if tc.expectErr {
				require.Error(t, err)
				if tc.errContains != "" {
					require.Contains(t, err.Error(), tc.errContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.Equal(t, tc.expectedPeriod, resp.StartingInfo.PreviousPeriod)
				require.Equal(t, tc.expectedHeight, resp.StartingInfo.Height)
				require.False(t, resp.StartingInfo.Stake.IsNegative())
			}
		})
	}
}
