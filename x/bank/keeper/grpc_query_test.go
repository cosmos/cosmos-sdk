// +build norace

package keeper_test

import (
	gocontext "context"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

func (suite *IntegrationTestSuite) TestQueryBalance() {
	app, ctx, queryClient := suite.app, suite.ctx, suite.queryClient
	_, _, addr := testdata.KeyTestPubAddr()

	_, err := queryClient.Balance(gocontext.Background(), &types.QueryBalanceRequest{})
	suite.Require().Error(err)

	_, err = queryClient.Balance(gocontext.Background(), &types.QueryBalanceRequest{Address: addr.String()})
	suite.Require().Error(err)

	req := types.NewQueryBalanceRequest(addr, fooDenom)
	res, err := queryClient.Balance(gocontext.Background(), req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.True(res.Balance.IsZero())

	origCoins := sdk.NewCoins(newFooCoin(50), newBarCoin(30))
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)

	app.AccountKeeper.SetAccount(ctx, acc)
	suite.Require().NoError(app.BankKeeper.SetBalances(ctx, acc.GetAddress(), origCoins))

	res, err = queryClient.Balance(gocontext.Background(), req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.True(res.Balance.IsEqual(newFooCoin(50)))
}

func (suite *IntegrationTestSuite) TestQueryAllBalances() {
	app, ctx, queryClient := suite.app, suite.ctx, suite.queryClient
	_, _, addr := testdata.KeyTestPubAddr()
	_, err := queryClient.AllBalances(gocontext.Background(), &types.QueryAllBalancesRequest{})
	suite.Require().Error(err)

	pageReq := &query.PageRequest{
		Key:        nil,
		Limit:      1,
		CountTotal: false,
	}
	req := types.NewQueryAllBalancesRequest(addr, pageReq)
	res, err := queryClient.AllBalances(gocontext.Background(), req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.True(res.Balances.IsZero())

	fooCoins := newFooCoin(50)
	barCoins := newBarCoin(30)

	origCoins := sdk.NewCoins(fooCoins, barCoins)
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)

	app.AccountKeeper.SetAccount(ctx, acc)
	suite.Require().NoError(app.BankKeeper.SetBalances(ctx, acc.GetAddress(), origCoins))

	res, err = queryClient.AllBalances(gocontext.Background(), req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Equal(res.Balances.Len(), 1)
	suite.NotNil(res.Pagination.NextKey)

	suite.T().Log("query second page with nextkey")
	pageReq = &query.PageRequest{
		Key:        res.Pagination.NextKey,
		Limit:      1,
		CountTotal: true,
	}
	req = types.NewQueryAllBalancesRequest(addr, pageReq)
	res, err = queryClient.AllBalances(gocontext.Background(), req)
	suite.Equal(res.Balances.Len(), 1)
	suite.Nil(res.Pagination.NextKey)
}

func (suite *IntegrationTestSuite) TestQueryTotalSupply() {
	app, ctx, queryClient := suite.app, suite.ctx, suite.queryClient
	expectedTotalSupply := types.NewSupply(sdk.NewCoins(sdk.NewInt64Coin("test", 400000000)))
	app.BankKeeper.SetSupply(ctx, expectedTotalSupply)

	res, err := queryClient.TotalSupply(gocontext.Background(), &types.QueryTotalSupplyRequest{})
	suite.Require().NoError(err)
	suite.Require().NotNil(res)

	suite.Require().Equal(expectedTotalSupply.Total, res.Supply)
}

func (suite *IntegrationTestSuite) TestQueryTotalSupplyOf() {
	app, ctx, queryClient := suite.app, suite.ctx, suite.queryClient

	test1Supply := sdk.NewInt64Coin("test1", 4000000)
	test2Supply := sdk.NewInt64Coin("test2", 700000000)
	expectedTotalSupply := types.NewSupply(sdk.NewCoins(test1Supply, test2Supply))
	app.BankKeeper.SetSupply(ctx, expectedTotalSupply)

	_, err := queryClient.SupplyOf(gocontext.Background(), &types.QuerySupplyOfRequest{})
	suite.Require().Error(err)

	res, err := queryClient.SupplyOf(gocontext.Background(), &types.QuerySupplyOfRequest{Denom: test1Supply.Denom})
	suite.Require().NoError(err)
	suite.Require().NotNil(res)

	suite.Require().Equal(test1Supply, res.Amount)
}

func (suite *IntegrationTestSuite) TestQueryParams() {
	res, err := suite.queryClient.Params(gocontext.Background(), &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().Equal(suite.app.BankKeeper.GetParams(suite.ctx), res.GetParams())
}
