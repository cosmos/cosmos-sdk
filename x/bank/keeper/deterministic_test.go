package keeper_test

import (
	"testing"

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
	encCfg      moduletestutil.TestEncodingConfig
	mintAcc     *authtypes.ModuleAccount
}

const (
	denomRegex = `[a-zA-Z][a-zA-Z0-9/:._-]{2,127}`
)

func TestDeterministicTestSuite(t *testing.T) {
	suite.Run(t, new(DeterministicTestSuite))
}

func (suite *DeterministicTestSuite) SetupTest() {
	key := sdk.NewKVStoreKey(banktypes.StoreKey)
	testCtx := testutil.DefaultContextWithDB(suite.T(), key, sdk.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithBlockHeader(tmproto.Header{Time: tmtime.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig(bank.AppModuleBasic{})

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

func (suite *DeterministicTestSuite) runQueryBalanceIterations(addr sdk.AccAddress, prevRes *sdk.Coin) {
	for i := 0; i < 1000; i++ {
		res, err := suite.queryClient.Balance(suite.ctx, banktypes.NewQueryBalanceRequest(addr, prevRes.GetDenom()))
		suite.Require().NoError(err)
		suite.Require().NotNil(res)

		suite.Require().Equal(res.GetBalance(), prevRes)
		prevRes = res.Balance
	}
}

func (suite *DeterministicTestSuite) TestGRPCQueryBalance() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		addr := testdata.AddressGenerator(t).Draw(t, "address")
		coin := sdk.NewCoin(
			rapid.StringMatching(denomRegex).Draw(t, "denom"),
			sdk.NewInt(rapid.Int64Min(1).Draw(t, "amount")),
		)

		suite.mockFundAccount(addr)
		err := banktestutil.FundAccount(suite.bankKeeper, suite.ctx, addr, sdk.NewCoins(coin))
		suite.Require().NoError(err)

		suite.runQueryBalanceIterations(addr, &coin)
	})

	addr := sdk.MustAccAddressFromBech32("cosmos139f7kncmglres2nf3h4hc4tade85ekfr8sulz5")
	coin := sdk.NewCoin(
		"denom",
		sdk.NewInt(10),
	)

	suite.mockFundAccount(addr)
	err := banktestutil.FundAccount(suite.bankKeeper, suite.ctx, addr, sdk.NewCoins(coin))
	suite.Require().NoError(err)

	suite.runQueryBalanceIterations(addr, &coin)
}

func (suite *DeterministicTestSuite) runAllBalancesIterations(addr sdk.AccAddress, prevRes sdk.Coins) {
	for i := 0; i < 1000; i++ {
		res, err := suite.queryClient.AllBalances(suite.ctx, &banktypes.QueryAllBalancesRequest{
			Address: addr.String(),
		})

		suite.Require().NoError(err)
		suite.Require().NotNil(res)
		suite.Require().NotNil(res.Balances)

		suite.Require().Equal(res.GetBalances(), prevRes)
		prevRes = res.GetBalances()
	}
}

func (suite *DeterministicTestSuite) TestGRPCQueryAllBalances() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		addr := testdata.AddressGenerator(t).Draw(t, "address")
		numCoins := rapid.IntRange(1, 10).Draw(t, "num-count")
		coins := make(sdk.Coins, 0, numCoins)

		for i := 0; i < numCoins; i++ {
			coin := sdk.NewCoin(
				rapid.StringMatching(denomRegex).Draw(t, "denom"),
				sdk.NewInt(rapid.Int64Min(1).Draw(t, "amount")),
			)

			// NewCoins sorts the denoms
			coins = sdk.NewCoins(append(coins, coin)...)
		}

		suite.mockFundAccount(addr)
		err := banktestutil.FundAccount(suite.bankKeeper, suite.ctx, addr, coins)
		suite.Require().NoError(err)

		suite.runAllBalancesIterations(addr, coins)
	})

	addr := sdk.MustAccAddressFromBech32("cosmos139f7kncmglres2nf3h4hc4tade85ekfr8sulz5")

	coins := sdk.NewCoins(
		sdk.NewCoin("stake", sdk.NewInt(10)),
		sdk.NewCoin("denom", sdk.NewInt(100)),
	)

	suite.mockFundAccount(addr)
	err := banktestutil.FundAccount(suite.bankKeeper, suite.ctx, addr, coins)
	suite.Require().NoError(err)

	suite.runAllBalancesIterations(addr, coins)
}

func (suite *DeterministicTestSuite) runSpendableBalancesIterations(addr sdk.AccAddress, prevRes sdk.Coins) {
	suite.authKeeper.EXPECT().GetAccount(suite.ctx, addr).Return(authtypes.NewBaseAccount(addr, nil, 10087, 0)).Times(1000)
	for i := 0; i < 1000; i++ {
		res, err := suite.queryClient.SpendableBalances(suite.ctx, &banktypes.QuerySpendableBalancesRequest{
			Address: addr.String(),
		})

		suite.Require().NoError(err)
		suite.Require().NotNil(res)
		suite.Require().NotNil(res.Balances)

		suite.Require().Equal(res.GetBalances(), prevRes)
		prevRes = res.GetBalances()
	}
}

func (suite *DeterministicTestSuite) TestGRPCQuerySpendableBalances() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		addr := testdata.AddressGenerator(t).Draw(t, "address")
		numCoins := rapid.IntRange(1, 10).Draw(t, "num-count")
		coins := make(sdk.Coins, 0, numCoins)

		for i := 0; i < numCoins; i++ {
			coin := sdk.NewCoin(
				rapid.StringMatching(denomRegex).Draw(t, "denom"),
				sdk.NewInt(rapid.Int64Min(1).Draw(t, "amount")),
			)

			// NewCoins sorts the denoms
			coins = sdk.NewCoins(append(coins, coin)...)
		}

		suite.mockFundAccount(addr)
		err := banktestutil.FundAccount(suite.bankKeeper, suite.ctx, addr, coins)
		suite.Require().NoError(err)

		suite.runSpendableBalancesIterations(addr, coins)
	})

	addr := sdk.MustAccAddressFromBech32("cosmos139f7kncmglres2nf3h4hc4tade85ekfr8sulz5")

	coins := sdk.NewCoins(
		sdk.NewCoin("stake", sdk.NewInt(10)),
		sdk.NewCoin("denom", sdk.NewInt(100)),
	)

	suite.mockFundAccount(addr)
	err := banktestutil.FundAccount(suite.bankKeeper, suite.ctx, addr, coins)
	suite.Require().NoError(err)

	suite.runSpendableBalancesIterations(addr, coins)
}
