package keeper_test

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/header"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	authtypes "cosmossdk.io/x/auth/types"
	poolkeeper "cosmossdk.io/x/protocolpool/keeper"
	pooltestutil "cosmossdk.io/x/protocolpool/testutil"
	"cosmossdk.io/x/protocolpool/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec/address"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

var (
	poolAcc   = authtypes.NewEmptyModuleAccount(types.ModuleName)
	streamAcc = authtypes.NewEmptyModuleAccount(types.StreamAccount)
)

type KeeperTestSuite struct {
	suite.Suite

	ctx           sdk.Context
	environment   appmodule.Environment
	poolKeeper    poolkeeper.Keeper
	authKeeper    *pooltestutil.MockAccountKeeper
	bankKeeper    *pooltestutil.MockBankKeeper
	stakingKeeper *pooltestutil.MockStakingKeeper

	msgServer   types.MsgServer
	queryServer types.QueryServer
}

func (s *KeeperTestSuite) SetupTest() {
	key := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	environment := runtime.NewEnvironment(storeService, log.NewNopLogger())
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithHeaderInfo(header.Info{Time: time.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{})

	// gomock initializations
	ctrl := gomock.NewController(s.T())
	accountKeeper := pooltestutil.NewMockAccountKeeper(ctrl)
	accountKeeper.EXPECT().GetModuleAddress(types.ModuleName).Return(poolAcc.GetAddress())
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()
	accountKeeper.EXPECT().GetModuleAddress(types.StreamAccount).Return(streamAcc.GetAddress())
	s.authKeeper = accountKeeper

	bankKeeper := pooltestutil.NewMockBankKeeper(ctrl)
	s.bankKeeper = bankKeeper

	stakingKeeper := pooltestutil.NewMockStakingKeeper(ctrl)
	stakingKeeper.EXPECT().BondDenom(ctx).Return("stake", nil).AnyTimes()
	s.stakingKeeper = stakingKeeper

	authority, err := accountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(types.GovModuleName))
	s.Require().NoError(err)

	poolKeeper := poolkeeper.NewKeeper(
		encCfg.Codec,
		environment,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		authority,
	)
	s.ctx = ctx
	s.poolKeeper = poolKeeper
	s.environment = environment

	types.RegisterInterfaces(encCfg.InterfaceRegistry)
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, encCfg.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, poolkeeper.Querier{Keeper: poolKeeper})
	s.msgServer = poolkeeper.NewMsgServerImpl(poolKeeper)
	s.queryServer = poolkeeper.NewQuerier(poolKeeper)
}

func (s *KeeperTestSuite) mockSendCoinsFromModuleToAccount(accAddr sdk.AccAddress) {
	s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(s.ctx, types.ModuleName, accAddr, gomock.Any()).AnyTimes()
}

func (s *KeeperTestSuite) mockWithdrawContinuousFund() {
	s.authKeeper.EXPECT().GetModuleAccount(s.ctx, types.ModuleName).Return(poolAcc).AnyTimes()
	distrBal := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100000)))
	s.bankKeeper.EXPECT().GetAllBalances(s.ctx, gomock.Any()).Return(distrBal).AnyTimes()
	s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(s.ctx, gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	s.stakingKeeper.EXPECT().BondDenom(s.ctx).AnyTimes()
}

func (s *KeeperTestSuite) mockStreamFunds() {
	s.authKeeper.EXPECT().GetModuleAccount(s.ctx, types.ModuleName).Return(poolAcc).AnyTimes()
	s.authKeeper.EXPECT().GetModuleAddress(types.StreamAccount).Return(streamAcc.GetAddress()).AnyTimes()
	distrBal := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100000)))
	s.bankKeeper.EXPECT().GetAllBalances(s.ctx, poolAcc.GetAddress()).Return(distrBal).AnyTimes()
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(s.ctx, poolAcc.GetName(), streamAcc.GetName(), gomock.Any()).AnyTimes()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
