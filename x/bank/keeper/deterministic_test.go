package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"
	"pgregory.net/rapid"
)

type DeterministicTestSuite struct {
	suite.Suite

	ctx        sdk.Context
	bankKeeper keeper.BaseKeeper
	authKeeper *banktestutil.MockAccountKeeper

	queryClient banktypes.QueryClient
	// msgServer   banktypes.MsgServer

	encCfg  moduletestutil.TestEncodingConfig
	mintAcc *authtypes.ModuleAccount
}

func (suite *DeterministicTestSuite) SetupTest() {
	key := sdk.NewKVStoreKey(banktypes.StoreKey)
	testCtx := testutil.DefaultContextWithDB(suite.T(), key, sdk.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithBlockHeader(tmproto.Header{Time: tmtime.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig(bank.AppModuleBasic{}, bank.AppModuleBasic{})

	// gomock initializations
	ctrl := gomock.NewController(suite.T())
	authKeeper := banktestutil.NewMockAccountKeeper(ctrl)

	suite.ctx = ctx
	suite.authKeeper = authKeeper
	suite.bankKeeper = keeper.NewBaseKeeper(
		encCfg.Codec,
		key,
		suite.authKeeper,
		map[string]bool{accAddrs[4].String(): true},
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// banktypes.RegisterInterfaces(encCfg.InterfaceRegistry)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, encCfg.InterfaceRegistry)
	banktypes.RegisterQueryServer(queryHelper, suite.bankKeeper)
	queryClient := banktypes.NewQueryClient(queryHelper)

	suite.queryClient = queryClient
	// suite.msgServer = keeper.NewMsgServerImpl(suite.bankKeeper)
	suite.encCfg = encCfg

	suite.mintAcc = authtypes.NewEmptyModuleAccount(minttypes.ModuleName, authtypes.Minter)
}

func (suite *DeterministicTestSuite) mockMintCoins(moduleAcc *authtypes.ModuleAccount) {
	suite.authKeeper.EXPECT().GetModuleAccount(suite.ctx, moduleAcc.Name).Return(moduleAcc)
}

func (suite *DeterministicTestSuite) mockSendCoinsFromModuleToAccount(moduleAcc *authtypes.ModuleAccount, accAddr sdk.AccAddress) {
	suite.authKeeper.EXPECT().GetModuleAddress(moduleAcc.Name).Return(moduleAcc.GetAddress())
	suite.authKeeper.EXPECT().GetAccount(suite.ctx, moduleAcc.GetAddress()).Return(moduleAcc)
	suite.authKeeper.EXPECT().HasAccount(suite.ctx, accAddr).Return(true)
}

func (suite *DeterministicTestSuite) mockFundAccount(receiver sdk.AccAddress) {
	suite.mockMintCoins(suite.mintAcc)
	suite.mockSendCoinsFromModuleToAccount(mintAcc, receiver)
}

func (suite *DeterministicTestSuite) runIterationsQueryBalance(addr sdk.AccAddress, prevRes *sdk.Coin) {
	for i := 0; i < 1000; i++ {
		res, err := suite.queryClient.Balance(suite.ctx, banktypes.NewQueryBalanceRequest(addr, prevRes.GetDenom()))
		suite.Require().NoError(err)
		suite.Require().NotNil(res)

		suite.Require().Equal(res.Balance, prevRes)
		prevRes = res.Balance
	}
}

func (suite *DeterministicTestSuite) TestGRPCQueryBalance() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		addr := testdata.AddressGenerator(t).Draw(t, "address")
		coin := sdk.NewCoin(
			rapid.StringMatching(`[A-Za-z]+[A-Za-z0-9]*`).Draw(t, "denom"),
			sdk.NewInt(rapid.Int64().Draw(t, "amount")),
		)

		suite.mockFundAccount(addr)
		suite.bankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, sdk.NewCoins(coin))
		suite.bankKeeper.SendCoinsFromModuleToAccount(suite.ctx, minttypes.ModuleName, addr, sdk.NewCoins(coin))

		suite.runIterationsQueryBalance(addr, &coin)
	})

	addr := sdk.MustAccAddressFromBech32("cosmos139f7kncmglres2nf3h4hc4tade85ekfr8sulz5")
	coin := sdk.NewCoin(
		"denom",
		sdk.NewInt(10),
	)

	suite.mockFundAccount(addr)
	suite.bankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, sdk.NewCoins(coin))
	suite.bankKeeper.SendCoinsFromModuleToAccount(suite.ctx, minttypes.ModuleName, addr, sdk.NewCoins(coin))

	suite.runIterationsQueryBalance(addr, &coin)
}
