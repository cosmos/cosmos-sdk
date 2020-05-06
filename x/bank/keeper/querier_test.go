package keeper_test

import (
	gocontext "context"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

func (suite *IntegrationTestSuite) TestQuerier_QueryBalance() {
	app, ctx := suite.app, suite.ctx
	_, _, addr := authtypes.KeyTestPubAddr()

	queryHelper := baseapp.NewQueryServerTestHelper(ctx)
	types.RegisterQueryServiceServer(queryHelper, keeper.Querier{app.BankKeeper})
	queryClient := types.NewQueryServiceClient(queryHelper)

	req := types.NewQueryBalanceRequest(addr, fooDenom)
	res, err := queryClient.QueryBalance(gocontext.Background(), req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.True(res.Balance.IsZero())

	origCoins := sdk.NewCoins(newFooCoin(50), newBarCoin(30))
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)

	app.AccountKeeper.SetAccount(ctx, acc)
	suite.Require().NoError(app.BankKeeper.SetBalances(ctx, acc.GetAddress(), origCoins))

	res, err = queryClient.QueryBalance(gocontext.Background(), req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.True(res.Balance.IsEqual(newFooCoin(50)))
}

func (suite *IntegrationTestSuite) TestQuerier_QueryAllBalances() {
	app, ctx := suite.app, suite.ctx
	_, _, addr := authtypes.KeyTestPubAddr()

	queryHelper := baseapp.NewQueryServerTestHelper(ctx)
	types.RegisterQueryServiceServer(queryHelper, keeper.Querier{app.BankKeeper})
	queryClient := types.NewQueryServiceClient(queryHelper)

	//res, err := queryClient.QueryAllBalances(nil, nil)
	//suite.Require().NotNil(err)
	//suite.Require().Nil(res)

	req := types.NewQueryAllBalancesRequest(addr)
	res, err := queryClient.QueryAllBalances(gocontext.Background(), req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.True(res.Balances.IsZero())

	origCoins := sdk.NewCoins(newFooCoin(50), newBarCoin(30))
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)

	app.AccountKeeper.SetAccount(ctx, acc)
	suite.Require().NoError(app.BankKeeper.SetBalances(ctx, acc.GetAddress(), origCoins))

	res, err = queryClient.QueryAllBalances(gocontext.Background(), req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().NoError(app.Codec().UnmarshalJSON(res, &balances))
	suite.True(balances.IsEqual(origCoins))
}

func (suite *IntegrationTestSuite) TestQuerier_QueryTotalSupply() {
	app, ctx := suite.app, suite.ctx
	expectedTotalSupply := bank.NewSupply(sdk.NewCoins(sdk.NewInt64Coin("test", 400000000)))
	app.BankKeeper.SetSupply(ctx, expectedTotalSupply)

	req := abci.RequestQuery{
		Path: fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryTotalSupply),
		Data: []byte{},
	}

	querier := keeper.NewQuerier(app.BankKeeper)

	res, err := querier(ctx, []string{types.QueryTotalSupply}, req)
	suite.Require().NotNil(err)
	suite.Require().Nil(res)

	req.Data = app.Codec().MustMarshalJSON(types.NewQueryTotalSupplyParams(1, 100))
	res, err = querier(ctx, []string{types.QueryTotalSupply}, req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)

	var resp sdk.Coins
	suite.Require().NoError(app.Codec().UnmarshalJSON(res, &resp))
	suite.Require().Equal(expectedTotalSupply.Total, resp)
}

func (suite *IntegrationTestSuite) TestQuerier_QueryTotalSupplyOf() {
	app, ctx := suite.app, suite.ctx

	test1Supply := sdk.NewInt64Coin("test1", 4000000)
	test2Supply := sdk.NewInt64Coin("test2", 700000000)
	expectedTotalSupply := bank.NewSupply(sdk.NewCoins(test1Supply, test2Supply))
	app.BankKeeper.SetSupply(ctx, expectedTotalSupply)

	req := abci.RequestQuery{
		Path: fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QuerySupplyOf),
		Data: []byte{},
	}

	querier := keeper.NewQuerier(app.BankKeeper)

	res, err := querier(ctx, []string{types.QuerySupplyOf}, req)
	suite.Require().NotNil(err)
	suite.Require().Nil(res)

	req.Data = app.Codec().MustMarshalJSON(types.NewQuerySupplyOfParams(test1Supply.Denom))
	res, err = querier(ctx, []string{types.QuerySupplyOf}, req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)

	var resp sdk.Int
	suite.Require().NoError(app.Codec().UnmarshalJSON(res, &resp))
	suite.Require().Equal(test1Supply.Amount, resp)
}

func (suite *IntegrationTestSuite) TestQuerierRouteNotFound() {
	app, ctx := suite.app, suite.ctx
	req := abci.RequestQuery{
		Path: fmt.Sprintf("custom/%s/invalid", types.ModuleName),
		Data: []byte{},
	}

	querier := keeper.NewQuerier(app.BankKeeper)
	_, err := querier(ctx, []string{"invalid"}, req)
	suite.Error(err)
	suite.True(res.Balances.IsEqual(origCoins))
}
