package keeper_test

import (
	"testing"

	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
	cmttime "github.com/cometbft/cometbft/types/time"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtestutil "github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	bondedAcc    = authtypes.NewEmptyModuleAccount(stakingtypes.BondedPoolName)
	notBondedAcc = authtypes.NewEmptyModuleAccount(stakingtypes.NotBondedPoolName)
	PKs          = simtestutil.CreateTestPubKeys(500)
)

type KeeperTestSuite struct {
	suite.Suite

	ctx           sdk.Context
	stakingKeeper *stakingkeeper.Keeper
	bankKeeper    *stakingtestutil.MockBankKeeper
	accountKeeper *stakingtestutil.MockAccountKeeper
	queryClient   stakingtypes.QueryClient
	msgServer     stakingtypes.MsgServer
}

func (s *KeeperTestSuite) SetupTest() {
	require := s.Require()
	key := storetypes.NewKVStoreKey(stakingtypes.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: cmttime.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig()

	ctrl := gomock.NewController(s.T())
	accountKeeper := stakingtestutil.NewMockAccountKeeper(ctrl)
	accountKeeper.EXPECT().GetModuleAddress(stakingtypes.BondedPoolName).Return(bondedAcc.GetAddress())
	accountKeeper.EXPECT().GetModuleAddress(stakingtypes.NotBondedPoolName).Return(notBondedAcc.GetAddress())
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
	require.NoError(keeper.SetParams(ctx, stakingtypes.DefaultParams()))

	s.ctx = ctx
	s.stakingKeeper = keeper
	s.bankKeeper = bankKeeper
	s.accountKeeper = accountKeeper

	stakingtypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, encCfg.InterfaceRegistry)
	stakingtypes.RegisterQueryServer(queryHelper, stakingkeeper.Querier{Keeper: keeper})
	s.queryClient = stakingtypes.NewQueryClient(queryHelper)
	s.msgServer = stakingkeeper.NewMsgServerImpl(keeper)
}

func (s *KeeperTestSuite) TestParams() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	expParams := stakingtypes.DefaultParams()
	// check that the empty keeper loads the default
	resParams, err := keeper.GetParams(ctx)
	require.NoError(err)
	require.Equal(expParams, resParams)

	expParams.MaxValidators = 555
	expParams.MaxEntries = 111
	require.NoError(keeper.SetParams(ctx, expParams))
	resParams, err = keeper.GetParams(ctx)
	require.NoError(err)
	require.True(expParams.Equal(resParams))
}

func (s *KeeperTestSuite) TestLastTotalPower() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	expTotalPower := math.NewInt(10 ^ 9)
	require.NoError(keeper.SetLastTotalPower(ctx, expTotalPower))
	resTotalPower, err := keeper.GetLastTotalPower(ctx)
	require.NoError(err)
	require.True(expTotalPower.Equal(resTotalPower))
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
