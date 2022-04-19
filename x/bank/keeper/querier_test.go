package keeper_test

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

func (suite *IntegrationTestSuite) TestQuerier_QueryBalance() {
	app, ctx := suite.app, suite.ctx
	legacyAmino := app.LegacyAmino()
	_, _, addr := testdata.KeyTestPubAddr()
	req := abci.RequestQuery{
		Path: fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryBalance),
		Data: []byte{},
	}

	querier := keeper.NewQuerier(app.BankKeeper, legacyAmino)

	res, err := querier(ctx, []string{types.QueryBalance}, req)
	suite.Require().NotNil(err)
	suite.Require().Nil(res)

	req.Data = legacyAmino.MustMarshalJSON(types.NewQueryBalanceRequest(addr, fooDenom))
	res, err = querier(ctx, []string{types.QueryBalance}, req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)

	var balance sdk.Coin
	suite.Require().NoError(legacyAmino.UnmarshalJSON(res, &balance))
	suite.True(balance.IsZero())

	origCoins := sdk.NewCoins(newFooCoin(50), newBarCoin(30))
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)

	app.AccountKeeper.SetAccount(ctx, acc)
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, acc.GetAddress(), origCoins))

	res, err = querier(ctx, []string{types.QueryBalance}, req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().NoError(legacyAmino.UnmarshalJSON(res, &balance))
	suite.True(balance.IsEqual(newFooCoin(50)))
}

func (suite *IntegrationTestSuite) TestQuerier_QueryAllBalances() {
	app, ctx := suite.app, suite.ctx
	legacyAmino := app.LegacyAmino()
	_, _, addr := testdata.KeyTestPubAddr()
	req := abci.RequestQuery{
		Path: fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryAllBalances),
		Data: []byte{},
	}

	querier := keeper.NewQuerier(app.BankKeeper, legacyAmino)

	res, err := querier(ctx, []string{types.QueryAllBalances}, req)
	suite.Require().NotNil(err)
	suite.Require().Nil(res)

	req.Data = legacyAmino.MustMarshalJSON(types.NewQueryAllBalancesRequest(addr, nil))
	res, err = querier(ctx, []string{types.QueryAllBalances}, req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)

	var balances sdk.Coins
	suite.Require().NoError(legacyAmino.UnmarshalJSON(res, &balances))
	suite.True(balances.IsZero())

	origCoins := sdk.NewCoins(newFooCoin(50), newBarCoin(30))
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)

	app.AccountKeeper.SetAccount(ctx, acc)
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, acc.GetAddress(), origCoins))
	res, err = querier(ctx, []string{types.QueryAllBalances}, req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().NoError(legacyAmino.UnmarshalJSON(res, &balances))
	suite.True(balances.IsEqual(origCoins))
}

func (suite *IntegrationTestSuite) TestQuerier_QueryTotalSupply() {
	app, ctx := suite.app, suite.ctx
	legacyAmino := app.LegacyAmino()

	genesisSupply, _, err := suite.app.BankKeeper.GetPaginatedTotalSupply(suite.ctx, &query.PageRequest{Limit: query.MaxLimit})
	suite.Require().NoError(err)

	testCoins := sdk.NewCoins(sdk.NewInt64Coin("test", 400000000))
	suite.
		Require().
		NoError(app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, testCoins))

	req := abci.RequestQuery{
		Path: fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryTotalSupply),
		Data: []byte{},
	}

	querier := keeper.NewQuerier(app.BankKeeper, legacyAmino)

	res, err := querier(ctx, []string{types.QueryTotalSupply}, req)
	suite.Require().NotNil(err)
	suite.Require().Nil(res)

	req.Data = legacyAmino.MustMarshalJSON(types.QueryTotalSupplyRequest{Pagination: &query.PageRequest{
		Limit: 100,
	}})
	res, err = querier(ctx, []string{types.QueryTotalSupply}, req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)

	var resp types.QueryTotalSupplyResponse
	suite.Require().NoError(legacyAmino.UnmarshalJSON(res, &resp))

	expectedTotalSupply := genesisSupply.Add(testCoins...)
	suite.Require().Equal(expectedTotalSupply, resp.Supply)
}

func (suite *IntegrationTestSuite) TestQuerier_QueryTotalSupplyOf() {
	app, ctx := suite.app, suite.ctx
	legacyAmino := app.LegacyAmino()
	test1Supply := sdk.NewInt64Coin("test1", 4000000)
	test2Supply := sdk.NewInt64Coin("test2", 700000000)
	expectedTotalSupply := sdk.NewCoins(test1Supply, test2Supply)
	suite.
		Require().
		NoError(app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, expectedTotalSupply))

	req := abci.RequestQuery{
		Path: fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QuerySupplyOf),
		Data: []byte{},
	}

	querier := keeper.NewQuerier(app.BankKeeper, legacyAmino)

	res, err := querier(ctx, []string{types.QuerySupplyOf}, req)
	suite.Require().NotNil(err)
	suite.Require().Nil(res)

	req.Data = legacyAmino.MustMarshalJSON(types.NewQuerySupplyOfParams(test1Supply.Denom))
	res, err = querier(ctx, []string{types.QuerySupplyOf}, req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)

	var resp sdk.Coin
	suite.Require().NoError(legacyAmino.UnmarshalJSON(res, &resp))
	suite.Require().Equal(test1Supply, resp)
}

func (suite *IntegrationTestSuite) TestQuerierRouteNotFound() {
	app, ctx := suite.app, suite.ctx
	legacyAmino := app.LegacyAmino()
	req := abci.RequestQuery{
		Path: fmt.Sprintf("custom/%s/invalid", types.ModuleName),
		Data: []byte{},
	}

	querier := keeper.NewQuerier(app.BankKeeper, legacyAmino)
	_, err := querier(ctx, []string{"invalid"}, req)
	suite.Error(err)
}
