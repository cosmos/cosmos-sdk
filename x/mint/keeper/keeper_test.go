package keeper_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	authtypes "cosmossdk.io/x/auth/types"
	"cosmossdk.io/x/mint"
	"cosmossdk.io/x/mint/keeper"
	minttestutil "cosmossdk.io/x/mint/testutil"
	"cosmossdk.io/x/mint/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

type IntegrationTestSuite struct {
	suite.Suite

	mintKeeper    keeper.Keeper
	ctx           sdk.Context
	stakingKeeper *minttestutil.MockStakingKeeper
	bankKeeper    *minttestutil.MockBankKeeper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupTest() {
	encCfg := moduletestutil.MakeTestEncodingConfig(mint.AppModuleBasic{})
	key := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	s.ctx = testCtx.Ctx

	// gomock initializations
	ctrl := gomock.NewController(s.T())
	accountKeeper := minttestutil.NewMockAccountKeeper(ctrl)
	bankKeeper := minttestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := minttestutil.NewMockStakingKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress(types.ModuleName).Return(sdk.AccAddress{})

	s.mintKeeper = keeper.NewKeeper(
		encCfg.Codec,
		storeService,
		stakingKeeper,
		accountKeeper,
		bankKeeper,
		authtypes.FeeCollectorName,
	)
	s.stakingKeeper = stakingKeeper
	s.bankKeeper = bankKeeper

	s.Require().Equal(testCtx.Ctx.Logger().With("module", "x/"+types.ModuleName),
		s.mintKeeper.Logger(testCtx.Ctx))

	s.Require().NoError(s.mintKeeper.Minter.Set(s.ctx, types.DefaultInitialMinter()))
}

func (s *IntegrationTestSuite) TestAliasFunctions() {
	stakingTokenSupply := math.NewIntFromUint64(100000000000)
	s.stakingKeeper.EXPECT().StakingTokenSupply(s.ctx).Return(stakingTokenSupply, nil)
	tokenSupply, err := s.mintKeeper.StakingTokenSupply(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal(tokenSupply, stakingTokenSupply)

	coins := sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(1000000)))
	s.bankKeeper.EXPECT().MintCoins(s.ctx, types.ModuleName, coins).Return(nil)
	s.Require().Equal(s.mintKeeper.MintCoins(s.ctx, sdk.NewCoins()), nil)
	s.Require().Nil(s.mintKeeper.MintCoins(s.ctx, coins))

	fees := sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(1000)))
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(s.ctx, types.ModuleName, authtypes.FeeCollectorName, fees).Return(nil)
	s.Require().Nil(s.mintKeeper.AddCollectedFees(s.ctx, fees))
}
