package keeper_test

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/header"
	storetypes "cosmossdk.io/store/types"
	authtypes "cosmossdk.io/x/auth/types"
	poolkeeper "cosmossdk.io/x/protocolpool/keeper"
	pooltestutil "cosmossdk.io/x/protocolpool/testutil"
	pooltypes "cosmossdk.io/x/protocolpool/types"

	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

var poolAcc = authtypes.NewEmptyModuleAccount(pooltypes.ModuleName)

type KeeperTestSuite struct {
	suite.Suite

	ctx        sdk.Context
	poolKeeper poolkeeper.Keeper
	bankKeeper *pooltestutil.MockBankKeeper
	msgServer  pooltypes.MsgServer
}

func (s *KeeperTestSuite) SetupTest() {
	key := storetypes.NewKVStoreKey(pooltypes.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithHeaderInfo(header.Info{Time: time.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig()

	// gomock initializations
	ctrl := gomock.NewController(s.T())
	accountKeeper := pooltestutil.NewMockAccountKeeper(ctrl)
	accountKeeper.EXPECT().GetModuleAddress(pooltypes.ModuleName).Return(poolAcc.GetAddress())
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

	bankKeeper := pooltestutil.NewMockBankKeeper(ctrl)
	s.bankKeeper = bankKeeper

	poolKeeper := poolkeeper.NewKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		bankKeeper,
		authtypes.NewModuleAddress(pooltypes.GovModuleName).String(),
	)
	s.ctx = ctx
	s.poolKeeper = poolKeeper

	pooltypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	s.msgServer = poolkeeper.NewMsgServerImpl(poolKeeper)
}

func (s *KeeperTestSuite) mockSendCoinsFromModuleToAccount(accAddr sdk.AccAddress) {
	s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(s.ctx, pooltypes.ModuleName, accAddr, gomock.Any()).AnyTimes()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
