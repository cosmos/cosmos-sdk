package staking_test

import (
	"math/rand"
	"testing"

	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
	cmttime "github.com/cometbft/cometbft/types/time"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtestutil "github.com/cosmos/cosmos-sdk/x/staking/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func BenchmarkApplyAndReturnValidatorSetUpdates(b *testing.B) {
	// goal of this benchmark to measure the performance changes in ApplyAndReturnValidatorSetUpdates
	// for dropping the bench32 conversion and different index types.
	// therefore the validator power, max or state is not modified to focus on comparing the valset
	// for an update only.
	const validatorCount = 150
	testEnv := newTestEnvironment(b)
	keeper, ctx := testEnv.stakingKeeper, testEnv.ctx
	r := rand.New(rand.NewSource(int64(1)))
	vals, valAddrs := setupState(b, r, validatorCount)

	params, err := keeper.GetParams(ctx)
	require.NoError(b, err)
	params.MaxValidators = uint32(validatorCount)
	require.NoError(b, keeper.SetParams(ctx, params))

	b.Logf("vals count: %d", validatorCount)
	for i, validator := range vals {
		require.NoError(b, keeper.SetValidator(ctx, validator))
		require.NoError(b, keeper.SetValidatorByConsAddr(ctx, validator))
		require.NoError(b, keeper.SetValidatorByPowerIndex(ctx, validator))
		require.NoError(b, keeper.SetLastValidatorPower(ctx, valAddrs[i], validator.ConsensusPower(sdk.DefaultPowerReduction)))
	}
	ctx, _ = testEnv.ctx.CacheContext()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	}
}

type KeeperTestEnvironment struct {
	ctx           sdk.Context
	stakingKeeper *stakingkeeper.Keeper
	bankKeeper    *stakingtestutil.MockBankKeeper
	accountKeeper *stakingtestutil.MockAccountKeeper
	queryClient   types.QueryClient
	msgServer     types.MsgServer
}

func newTestEnvironment(tb testing.TB) *KeeperTestEnvironment {
	tb.Helper()
	key := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(tb, key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: cmttime.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig()

	ctrl := gomock.NewController(tb)
	accountKeeper := stakingtestutil.NewMockAccountKeeper(ctrl)
	accountKeeper.EXPECT().GetModuleAddress(types.BondedPoolName).
		Return(authtypes.NewEmptyModuleAccount(types.BondedPoolName).GetAddress())
	accountKeeper.EXPECT().GetModuleAddress(types.NotBondedPoolName).
		Return(authtypes.NewEmptyModuleAccount(types.NotBondedPoolName).GetAddress())
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

	bankKeeper := stakingtestutil.NewMockBankKeeper(ctrl)

	keeper := stakingkeeper.NewKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		bankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		address.NewBech32Codec("cosmosvaloper"),
		address.NewBech32Codec("cosmosvalcons"),
	)
	require.NoError(tb, keeper.SetParams(ctx, types.DefaultParams()))

	testEnv := &KeeperTestEnvironment{
		ctx:           ctx,
		stakingKeeper: keeper,
		bankKeeper:    bankKeeper,
		accountKeeper: accountKeeper,
	}
	types.RegisterInterfaces(encCfg.InterfaceRegistry)
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, encCfg.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, stakingkeeper.Querier{Keeper: keeper})
	testEnv.queryClient = types.NewQueryClient(queryHelper)
	testEnv.msgServer = stakingkeeper.NewMsgServerImpl(keeper)
	return testEnv
}

func setupState(b *testing.B, r *rand.Rand, numBonded int) ([]types.Validator, []sdk.ValAddress) {
	b.Helper()
	accs := simtypes.RandomAccounts(r, numBonded)
	initialStake := sdkmath.NewInt(r.Int63n(1000) + 10)

	validators := make([]types.Validator, numBonded)
	valAddrs := make([]sdk.ValAddress, numBonded)

	for i := 0; i < numBonded; i++ {
		valAddr := sdk.ValAddress(accs[i].Address)
		valAddrs[i] = valAddr

		maxCommission := sdkmath.LegacyNewDecWithPrec(int64(simtypes.RandIntBetween(r, 1, 100)), 2)
		commission := types.NewCommission(
			simtypes.RandomDecAmount(r, maxCommission),
			maxCommission,
			simtypes.RandomDecAmount(r, maxCommission),
		)

		validator, err := types.NewValidator(valAddr.String(), accs[i].ConsKey.PubKey(), types.Description{})
		require.NoError(b, err)
		startStake := sdkmath.NewInt(r.Int63n(1000) + initialStake.Int64())
		validator.Tokens = startStake.Mul(sdk.DefaultPowerReduction)
		validator.DelegatorShares = sdkmath.LegacyNewDecFromInt(initialStake)
		validator.Commission = commission
		validator.Status = types.Bonded
		validators[i] = validator
	}
	return validators, valAddrs
}
