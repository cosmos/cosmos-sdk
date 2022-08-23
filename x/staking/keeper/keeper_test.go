package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/testutil"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	stakingtestutil "github.com/cosmos/cosmos-sdk/x/staking/testutil"
)

var (
	bondedAcc    = authtypes.NewEmptyModuleAccount(stakingtypes.BondedPoolName)
	notBondedAcc = authtypes.NewEmptyModuleAccount(stakingtypes.NotBondedPoolName)
	PKs          = simtestutil.CreateTestPubKeys(500)
)

type KeeperTestSuite struct {
	suite.Suite

	ctx           sdk.Context
	stakingKeeper stakingkeeper.Keeper
	bankKeeper    *stakingtestutil.MockBankKeeper
}

func (suite *KeeperTestSuite) SetupTest() {
	key := sdk.NewKVStoreKey(stakingtypes.StoreKey)
	testCtx := testutil.DefaultContextWithDB(suite.T(), key, sdk.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithBlockHeader(tmproto.Header{Time: tmtime.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig()

	ctrl := gomock.NewController(suite.T())
	accountKeeper := stakingtestutil.NewMockAccountKeeper(ctrl)
	accountKeeper.EXPECT().GetModuleAddress(stakingtypes.BondedPoolName).Return(bondedAcc.GetAddress())
	accountKeeper.EXPECT().GetModuleAddress(stakingtypes.NotBondedPoolName).Return(notBondedAcc.GetAddress())
	bankKeeper := stakingtestutil.NewMockBankKeeper(ctrl)

	keeper := *stakingkeeper.NewKeeper(
		encCfg.Codec,
		key,
		accountKeeper,
		bankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	keeper.SetParams(ctx, stakingtypes.DefaultParams())

	suite.ctx = ctx
	suite.stakingKeeper = keeper
	suite.bankKeeper = bankKeeper
}

func (suite *KeeperTestSuite) TestParams() {
	ctx, keeper := suite.ctx, suite.stakingKeeper
	require := suite.Require()

	expParams := stakingtypes.DefaultParams()
	expParams.MaxValidators = 555
	expParams.MaxEntries = 111
	keeper.SetParams(ctx, expParams)
	resParams := keeper.GetParams(ctx)
	require.True(expParams.Equal(resParams))
}

func (suite *KeeperTestSuite) TestLastTotalPower() {
	ctx, keeper := suite.ctx, suite.stakingKeeper
	require := suite.Require()

	expTotalPower := math.NewInt(10 ^ 9)
	keeper.SetLastTotalPower(ctx, expTotalPower)
	resTotalPower := keeper.GetLastTotalPower(ctx)
	require.True(expTotalPower.Equal(resTotalPower))
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
