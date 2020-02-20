package keeper_test

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

func (suite *IntegrationTestSuite) TestQuerier_QueryBalance() {
	app, ctx := suite.app, suite.ctx
	_, _, addr := authtypes.KeyTestPubAddr()
	req := abci.RequestQuery{
		Path: fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryBalance),
		Data: []byte{},
	}

	querier := keeper.NewQuerier(app.BankKeeper)

	res, err := querier(ctx, []string{types.QueryBalance}, req)
	suite.Require().NotNil(err)
	suite.Require().Nil(res)

	req.Data = app.Codec().MustMarshalJSON(types.NewQueryBalanceParams(addr, fooDenom))
	res, err = querier(ctx, []string{types.QueryBalance}, req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)

	var balance sdk.Coin
	suite.Require().NoError(app.Codec().UnmarshalJSON(res, &balance))
	suite.True(balance.IsZero())

	origCoins := sdk.NewCoins(newFooCoin(50), newBarCoin(30))
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)

	app.AccountKeeper.SetAccount(ctx, acc)
	suite.Require().NoError(app.BankKeeper.SetBalances(ctx, acc.GetAddress(), origCoins))

	res, err = querier(ctx, []string{types.QueryBalance}, req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().NoError(app.Codec().UnmarshalJSON(res, &balance))
	suite.True(balance.IsEqual(newFooCoin(50)))
}

func (suite *IntegrationTestSuite) TestQuerier_QueryAllBalances() {
	app, ctx := suite.app, suite.ctx
	_, _, addr := authtypes.KeyTestPubAddr()
	req := abci.RequestQuery{
		Path: fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryAllBalances),
		Data: []byte{},
	}

	querier := keeper.NewQuerier(app.BankKeeper)

	res, err := querier(ctx, []string{types.QueryAllBalances}, req)
	suite.Require().NotNil(err)
	suite.Require().Nil(res)

	req.Data = app.Codec().MustMarshalJSON(types.NewQueryAllBalancesParams(addr))
	res, err = querier(ctx, []string{types.QueryAllBalances}, req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)

	var balances sdk.Coins
	suite.Require().NoError(app.Codec().UnmarshalJSON(res, &balances))
	suite.True(balances.IsZero())

	origCoins := sdk.NewCoins(newFooCoin(50), newBarCoin(30))
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)

	app.AccountKeeper.SetAccount(ctx, acc)
	suite.Require().NoError(app.BankKeeper.SetBalances(ctx, acc.GetAddress(), origCoins))

	res, err = querier(ctx, []string{types.QueryAllBalances}, req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().NoError(app.Codec().UnmarshalJSON(res, &balances))
	suite.True(balances.IsEqual(origCoins))
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
}
