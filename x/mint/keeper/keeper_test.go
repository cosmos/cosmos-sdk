package keeper_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttestutil "github.com/cosmos/cosmos-sdk/x/mint/testutil"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	mintKeeper    keeper.Keeper
	ctx           sdk.Context
	msgServer     types.MsgServer
	stakingKeeper *minttestutil.MockStakingKeeper
	bankKeeper    *minttestutil.MockBankKeeper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupTest() {
	encCfg := moduletestutil.MakeTestEncodingConfig(mint.AppModuleBasic{})
	key := storetypes.NewKVStoreKey(types.StoreKey)
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
		key,
		stakingKeeper,
		accountKeeper,
		bankKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	s.stakingKeeper = stakingKeeper
	s.bankKeeper = bankKeeper

	s.Require().Equal(testCtx.Ctx.Logger().With("module", "x/"+types.ModuleName),
		s.mintKeeper.Logger(testCtx.Ctx))

	err := s.mintKeeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)
	s.mintKeeper.SetMinter(s.ctx, types.DefaultInitialMinter())

	s.msgServer = keeper.NewMsgServerImpl(s.mintKeeper)
}

func (s *IntegrationTestSuite) TestParams() {
	testCases := []struct {
		name      string
		input     types.Params
		expectErr bool
	}{
		{
			name: "set invalid params (⚠️ not validated in keeper)",
			input: types.Params{
				MintDenom:           sdk.DefaultBondDenom,
				InflationRateChange: sdkmath.LegacyNewDecWithPrec(-13, 2),
				InflationMax:        sdkmath.LegacyNewDecWithPrec(20, 2),
				InflationMin:        sdkmath.LegacyNewDecWithPrec(7, 2),
				GoalBonded:          sdkmath.LegacyNewDecWithPrec(67, 2),
				BlocksPerYear:       uint64(60 * 60 * 8766 / 5),
			},
			expectErr: false,
		},
		{
			name: "set full valid params",
			input: types.Params{
				MintDenom:           sdk.DefaultBondDenom,
				InflationRateChange: sdkmath.LegacyNewDecWithPrec(8, 2),
				InflationMax:        sdkmath.LegacyNewDecWithPrec(20, 2),
				InflationMin:        sdkmath.LegacyNewDecWithPrec(2, 2),
				GoalBonded:          sdkmath.LegacyNewDecWithPrec(37, 2),
				BlocksPerYear:       uint64(60 * 60 * 8766 / 5),
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			expected := s.mintKeeper.GetParams(s.ctx)
			err := s.mintKeeper.SetParams(s.ctx, tc.input)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				expected = tc.input
				s.Require().NoError(err)
			}

			p := s.mintKeeper.GetParams(s.ctx)
			s.Require().Equal(expected, p)
		})
	}
}

func (s *IntegrationTestSuite) TestAliasFunctions() {
	stakingTokenSupply := sdk.NewIntFromUint64(100000000000)
	s.stakingKeeper.EXPECT().StakingTokenSupply(s.ctx).Return(stakingTokenSupply)
	s.Require().Equal(s.mintKeeper.StakingTokenSupply(s.ctx), stakingTokenSupply)

	bondedRatio := sdkmath.LegacyNewDecWithPrec(15, 2)
	s.stakingKeeper.EXPECT().BondedRatio(s.ctx).Return(bondedRatio)
	s.Require().Equal(s.mintKeeper.BondedRatio(s.ctx), bondedRatio)

	coins := sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1000000)))
	s.bankKeeper.EXPECT().MintCoins(s.ctx, types.ModuleName, coins).Return(nil)
	s.Require().Equal(s.mintKeeper.MintCoins(s.ctx, sdk.NewCoins()), nil)
	s.Require().Nil(s.mintKeeper.MintCoins(s.ctx, coins))

	fees := sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1000)))
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(s.ctx, types.ModuleName, authtypes.FeeCollectorName, fees).Return(nil)
	s.Require().Nil(s.mintKeeper.AddCollectedFees(s.ctx, fees))
}
