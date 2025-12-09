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

func TestQueryValidatorHistoricalRewards(t *testing.T) {
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
	stakingKeeper.EXPECT().Delegation(gomock.Any(), addr, valAddr).Return(del, nil).AnyTimes()

	err = distrtestutil.CallCreateValidatorHooks(ctx, distrKeeper, addr, valAddr)
	require.NoError(t, err)

	tokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(100)}}
	require.NoError(t, distrKeeper.AllocateTokensToValidator(ctx, val, tokens))
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	_, err = distrKeeper.IncrementValidatorPeriod(ctx, val)
	require.NoError(t, err)

	tokens2 := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(200)}}
	require.NoError(t, distrKeeper.AllocateTokensToValidator(ctx, val, tokens2))
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	_, err = distrKeeper.IncrementValidatorPeriod(ctx, val)
	require.NoError(t, err)

	querier := keeper.NewQuerier(distrKeeper)

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
				ValidatorAddress: valAddr.String(),
				Period:           1,
			},
			expectErr:          false,
			expectedRefCount:   1,
			expectNonZeroRatio: false,
		},
		{
			name: "period 2 deleted (reference count 0)",
			req: &disttypes.QueryValidatorHistoricalRewardsRequest{
				ValidatorAddress: valAddr.String(),
				Period:           2,
			},
			expectErr:          false,
			expectedRefCount:   0,
			expectNonZeroRatio: false,
		},
		{
			name: "period 3 is current with rewards",
			req: &disttypes.QueryValidatorHistoricalRewardsRequest{
				ValidatorAddress: valAddr.String(),
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
				ValidatorAddress: valAddr.String(),
				Period:           999,
			},
			expectErr:          false,
			expectedRefCount:   0,
			expectNonZeroRatio: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := querier.ValidatorHistoricalRewards(ctx, tc.req)
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
