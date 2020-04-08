package keeper_test

import (
	gocontext "context"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

func (suite *IntegrationTestSuite) TestQuerier_QueryBalance() {
	app, ctx := suite.app, suite.ctx
	_, _, addr := authtypes.KeyTestPubAddr()

	queryHelper := baseapp.NewQueryServerTestHelper(ctx)
	types.RegisterQueryServer(queryHelper, keeper.Querier{app.BankKeeper})
	queryClient := types.NewQueryClient(queryHelper)

	req := types.NewQueryBalanceRequest(addr, fooDenom)
	balance, err := queryClient.QueryBalance(gocontext.Background(), &req)
	suite.Require().NoError(err)
	suite.Require().NotNil(balance)
	suite.True(balance.IsZero())

	origCoins := sdk.NewCoins(newFooCoin(50), newBarCoin(30))
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)

	app.AccountKeeper.SetAccount(ctx, acc)
	suite.Require().NoError(app.BankKeeper.SetBalances(ctx, acc.GetAddress(), origCoins))

	balance, err = queryClient.QueryBalance(gocontext.Background(), &req)
	suite.Require().NoError(err)
	suite.Require().NotNil(balance)
	suite.True(balance.IsEqual(newFooCoin(50)))
}

func (suite *IntegrationTestSuite) TestQuerier_QueryAllBalances() {
	app, ctx := suite.app, suite.ctx
	_, _, addr := authtypes.KeyTestPubAddr()

	queryHelper := baseapp.NewQueryServerTestHelper(ctx)
	types.RegisterQueryServer(queryHelper, keeper.Querier{app.BankKeeper})
	queryClient := types.NewQueryClient(queryHelper)

	//res, err := queryClient.QueryAllBalances(nil, nil)
	//suite.Require().NotNil(err)
	//suite.Require().Nil(res)

	req := types.NewQueryAllBalancesRequest(addr)
	res, err := queryClient.QueryAllBalances(gocontext.Background(), &req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)

	suite.True(res.Balances.IsZero())

	origCoins := sdk.NewCoins(newFooCoin(50), newBarCoin(30))
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)

	app.AccountKeeper.SetAccount(ctx, acc)
	suite.Require().NoError(app.BankKeeper.SetBalances(ctx, acc.GetAddress(), origCoins))

	res, err = queryClient.QueryAllBalances(gocontext.Background(), &req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.True(res.Balances.IsEqual(origCoins))
}
