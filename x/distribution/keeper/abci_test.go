package keeper_test

import (
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
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

var (
	testProposerAddress = sdk.ConsAddress("test")
)

type testSetup struct {
	testCtx       testutil.TestContext
	bankKeeper    *distrtestutil.MockBankKeeper
	stakingKeeper *distrtestutil.MockStakingKeeper
	accountKeeper *distrtestutil.MockAccountKeeper
	distrKeeper   keeper.Keeper
}

func setupTest(t *testing.T, protocolPoolEnabled bool) testSetup {
	t.Helper()

	ctrl := gomock.NewController(t)
	key := storetypes.NewKVStoreKey(disttypes.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress(disttypes.ModuleName).Return(distrAcc.GetAddress()).Times(1)

	var opts []keeper.InitOptions
	if protocolPoolEnabled {
		opts = append(opts, keeper.WithProtocolPoolEnabled())
	}

	distrKeeper := keeper.NewKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		"fee_collector",
		authtypes.NewModuleAddress("gov").String(),
		opts...,
	)

	return testSetup{
		testCtx:       testCtx,
		bankKeeper:    bankKeeper,
		distrKeeper:   distrKeeper,
		stakingKeeper: stakingKeeper,
		accountKeeper: accountKeeper,
	}
}

func TestBeginBlockNoOp(t *testing.T) {
	ts := setupTest(t)
	ctx := ts.testCtx.Ctx.
		WithBlockHeader(cmtproto.Header{
			ProposerAddress: testProposerAddress,
			Time:            time.Now(),
		}).
		WithBlockHeight(0)

	err := ts.distrKeeper.BeginBlocker(ctx)
	require.NoError(t, err)

	// check cons address
	got, err := ts.distrKeeper.GetPreviousProposerConsAddr(ctx)
	require.NoError(t, err)
	require.Equal(t, testProposerAddress, got)
}

func TestBeginBlockToManyValidators(t *testing.T) {
	ts := setupTest(t)

	ctx := ts.testCtx.Ctx.
		WithBlockHeader(cmtproto.Header{
			ProposerAddress: testProposerAddress,
			Time:            time.Now(),
		})

	feeCollectorAcc := authtypes.NewEmptyModuleAccount("fee_collector")
	ts.accountKeeper.EXPECT().GetModuleAccount(gomock.Any(), "fee_collector").Return(feeCollectorAcc)
	ts.stakingKeeper.EXPECT().ValidatorAddressCodec().Return(address.NewBech32Codec("cosmosvaloper")).AnyTimes()

	// reset fee pool & set params
	require.NoError(t, ts.distrKeeper.Params.Set(ctx, disttypes.DefaultParams()))
	require.NoError(t, ts.distrKeeper.FeePool.Set(ctx, disttypes.InitialFeePool()))

	// create validator with 50% commission
	valAddr0 := sdk.ValAddress(valConsAddr0)
	val0, err := distrtestutil.CreateValidator(valConsPk0, math.NewInt(100))
	require.NoError(t, err)
	val0.Commission = stakingtypes.NewCommission(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0))
	ts.stakingKeeper.EXPECT().ValidatorByConsAddr(gomock.Any(), sdk.GetConsAddress(valConsPk0)).Return(val0, nil).AnyTimes()

	// create second validator with 0% commission
	valAddr1 := sdk.ValAddress(valConsAddr1)
	val1, err := distrtestutil.CreateValidator(valConsPk1, math.NewInt(100))
	require.NoError(t, err)
	val1.Commission = stakingtypes.NewCommission(math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0))
	ts.stakingKeeper.EXPECT().ValidatorByConsAddr(gomock.Any(), sdk.GetConsAddress(valConsPk1)).Return(val1, nil).AnyTimes()

	// assert initial state: zero outstanding rewards, zero community pool, zero commission, zero current rewards
	val0OutstandingRewards, err := ts.distrKeeper.GetValidatorOutstandingRewards(ctx, valAddr0)
	require.NoError(t, err)
	require.True(t, val0OutstandingRewards.Rewards.IsZero())

	val1OutstandingRewards, err := ts.distrKeeper.GetValidatorOutstandingRewards(ctx, valAddr1)
	require.NoError(t, err)
	require.True(t, val1OutstandingRewards.Rewards.IsZero())

	feePool, err := ts.distrKeeper.FeePool.Get(ctx)
	require.NoError(t, err)
	require.True(t, feePool.CommunityPool.IsZero())

	val0Commission, err := ts.distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddr0)
	require.NoError(t, err)
	require.True(t, val0Commission.Commission.IsZero())

	val1Commission, err := ts.distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddr1)
	require.NoError(t, err)
	require.True(t, val1Commission.Commission.IsZero())

	val0CurrentRewards, err := ts.distrKeeper.GetValidatorCurrentRewards(ctx, valAddr0)
	require.NoError(t, err)
	require.True(t, val0CurrentRewards.Rewards.IsZero())

	val1CurrentRewards, err := ts.distrKeeper.GetValidatorCurrentRewards(ctx, valAddr1)
	require.NoError(t, err)
	require.True(t, val1CurrentRewards.Rewards.IsZero())

	// allocate tokens as if both had voted and second was proposer
	fees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100)))
	ts.bankKeeper.EXPECT().GetAllBalances(gomock.Any(), feeCollectorAcc.GetAddress()).Return(fees)
	ts.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), "fee_collector", disttypes.ModuleName, fees)

	votes := []abci.VoteInfo{
		{
			Validator: abci.Validator{
				Address: valConsPk0.Address(),
				Power:   100,
			},
		},
		{
			Validator: abci.Validator{
				Address: valConsPk1.Address(),
				Power:   100,
			},
		},
	}
	ctx = ctx.WithVoteInfos(votes).WithBlockHeight(2)

	require.NoError(t, ts.distrKeeper.BeginBlocker(ctx))

	// 98 outstanding rewards (100 less 2 to community pool)
	val0OutstandingRewards, err = ts.distrKeeper.GetValidatorOutstandingRewards(ctx, valAddr0)
	require.NoError(t, err)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDecWithPrec(490, 1)}}, val0OutstandingRewards.Rewards)

	val1OutstandingRewards, err = ts.distrKeeper.GetValidatorOutstandingRewards(ctx, valAddr1)
	require.NoError(t, err)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDecWithPrec(490, 1)}}, val1OutstandingRewards.Rewards)

	// 2 community pool coins
	feePool, err = ts.distrKeeper.FeePool.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(2)}}, feePool.CommunityPool)

	// 50% commission for first proposer, (0.5 * 98%) * 100 / 2 = 23.25
	val0Commission, err = ts.distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddr0)
	require.NoError(t, err)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDecWithPrec(2450, 2)}}, val0Commission.Commission)

	// zero commission for second proposer
	val1Commission, err = ts.distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddr1)
	require.NoError(t, err)
	require.True(t, val1Commission.Commission.IsZero())

	// just staking.proportional for first proposer less commission = (0.5 * 98%) * 100 / 2 = 24.50
	val0CurrentRewards, err = ts.distrKeeper.GetValidatorCurrentRewards(ctx, valAddr0)
	require.NoError(t, err)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDecWithPrec(2450, 2)}}, val0CurrentRewards.Rewards)

	// proposer reward + staking.proportional for second proposer = (0.5 * (98%)) * 100 = 49
	val1CurrentRewards, err = ts.distrKeeper.GetValidatorCurrentRewards(ctx, valAddr1)
	require.NoError(t, err)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDecWithPrec(490, 1)}}, val1CurrentRewards.Rewards)

	// check cons address
	got, err := ts.distrKeeper.GetPreviousProposerConsAddr(ctx)
	require.NoError(t, err)
	require.Equal(t, testProposerAddress, got)
}

func TestBeginBlockTruncation(t *testing.T) {
	ts := setupTest(t)
	ctx := ts.testCtx.Ctx.
		WithBlockHeader(cmtproto.Header{
			ProposerAddress: testProposerAddress,
			Time:            time.Now(),
		})

	feeCollectorAcc := authtypes.NewEmptyModuleAccount("fee_collector")
	ts.accountKeeper.EXPECT().GetModuleAccount(gomock.Any(), "fee_collector").Return(feeCollectorAcc)
	ts.stakingKeeper.EXPECT().ValidatorAddressCodec().Return(address.NewBech32Codec("cosmosvaloper")).AnyTimes()

	// reset fee pool
	require.NoError(t, ts.distrKeeper.FeePool.Set(ctx, disttypes.InitialFeePool()))
	require.NoError(t, ts.distrKeeper.Params.Set(ctx, disttypes.DefaultParams()))

	// create validator with 10% commission
	valAddr0 := sdk.ValAddress(valConsAddr0)
	val0, err := distrtestutil.CreateValidator(valConsPk0, math.NewInt(100))
	require.NoError(t, err)
	val0.Commission = stakingtypes.NewCommission(math.LegacyNewDecWithPrec(1, 1), math.LegacyNewDecWithPrec(1, 1), math.LegacyNewDec(0))
	ts.stakingKeeper.EXPECT().ValidatorByConsAddr(gomock.Any(), sdk.GetConsAddress(valConsPk0)).Return(val0, nil).AnyTimes()

	// create second validator with 10% commission
	valAddr1 := sdk.ValAddress(valConsAddr1)
	val1, err := distrtestutil.CreateValidator(valConsPk1, math.NewInt(100))
	require.NoError(t, err)
	val1.Commission = stakingtypes.NewCommission(math.LegacyNewDecWithPrec(1, 1), math.LegacyNewDecWithPrec(1, 1), math.LegacyNewDec(0))
	ts.stakingKeeper.EXPECT().ValidatorByConsAddr(gomock.Any(), sdk.GetConsAddress(valConsPk1)).Return(val1, nil).AnyTimes()

	// create third validator with 10% commission
	valAddr2 := sdk.ValAddress(valConsAddr2)
	val2, err := stakingtypes.NewValidator(sdk.ValAddress(valConsAddr2).String(), valConsPk1, stakingtypes.Description{})
	require.NoError(t, err)
	val2.Commission = stakingtypes.NewCommission(math.LegacyNewDecWithPrec(1, 1), math.LegacyNewDecWithPrec(1, 1), math.LegacyNewDec(0))
	ts.stakingKeeper.EXPECT().ValidatorByConsAddr(gomock.Any(), sdk.GetConsAddress(valConsPk2)).Return(val2, nil).AnyTimes()

	// assert initial state: zero outstanding rewards, zero community pool, zero commission, zero current rewards
	val0OutstandingRewards, err := ts.distrKeeper.GetValidatorOutstandingRewards(ctx, valAddr0)
	require.NoError(t, err)
	require.True(t, val0OutstandingRewards.Rewards.IsZero())

	val1OutstandingRewards, err := ts.distrKeeper.GetValidatorOutstandingRewards(ctx, valAddr1)
	require.NoError(t, err)
	require.True(t, val1OutstandingRewards.Rewards.IsZero())

	feePool, err := ts.distrKeeper.FeePool.Get(ctx)
	require.NoError(t, err)
	require.True(t, feePool.CommunityPool.IsZero())

	val0Commission, err := ts.distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddr0)
	require.NoError(t, err)
	require.True(t, val0Commission.Commission.IsZero())

	val1Commission, err := ts.distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddr1)
	require.NoError(t, err)
	require.True(t, val1Commission.Commission.IsZero())

	val0CurrentRewards, err := ts.distrKeeper.GetValidatorCurrentRewards(ctx, valAddr0)
	require.NoError(t, err)
	require.True(t, val0CurrentRewards.Rewards.IsZero())

	val1CurrentRewards, err := ts.distrKeeper.GetValidatorCurrentRewards(ctx, valAddr1)
	require.NoError(t, err)
	require.True(t, val1CurrentRewards.Rewards.IsZero())

	// allocate tokens as if both had voted and second was proposer
	fees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(634195840)))
	ts.bankKeeper.EXPECT().GetAllBalances(gomock.Any(), feeCollectorAcc.GetAddress()).Return(fees)
	ts.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), "fee_collector", disttypes.ModuleName, fees)

	votes := []abci.VoteInfo{
		{
			Validator: abci.Validator{
				Address: valConsPk0.Address(),
				Power:   11,
			},
		},
		{
			Validator: abci.Validator{
				Address: valConsPk1.Address(),
				Power:   10,
			},
		},
		{
			Validator: abci.Validator{
				Address: valConsPk2.Address(),
				Power:   10,
			},
		},
	}
	ctx = ctx.WithVoteInfos(votes).WithBlockHeight(2)
	require.NoError(t, ts.distrKeeper.BeginBlocker(ctx))

	val0OutstandingRewards, err = ts.distrKeeper.GetValidatorOutstandingRewards(ctx, valAddr0)
	require.NoError(t, err)
	require.True(t, val0OutstandingRewards.Rewards.IsValid())

	val1OutstandingRewards, err = ts.distrKeeper.GetValidatorOutstandingRewards(ctx, valAddr1)
	require.NoError(t, err)
	require.True(t, val1OutstandingRewards.Rewards.IsValid())

	val2OutstandingRewards, err := ts.distrKeeper.GetValidatorOutstandingRewards(ctx, valAddr2)
	require.NoError(t, err)
	require.True(t, val2OutstandingRewards.Rewards.IsValid())

	// check cons address
	got, err := ts.distrKeeper.GetPreviousProposerConsAddr(ctx)
	require.NoError(t, err)
	require.Equal(t, testProposerAddress, got)
}
