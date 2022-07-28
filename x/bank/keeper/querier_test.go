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

func (suite *KeeperTestSuite) TestQuerier_QueryBalance() {
	ctx, legacyAmino := suite.ctx, suite.encCfg.Amino
	_, _, addr := testdata.KeyTestPubAddr()
	req := abci.RequestQuery{
		Path: fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryBalance),
		Data: []byte{},
	}

	querier := keeper.NewQuerier(suite.bankKeeper, legacyAmino)

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

	suite.mockFundAccount(addr)
	suite.Require().NoError(testutil.FundAccount(suite.bankKeeper, ctx, addr, origCoins))

	res, err = querier(ctx, []string{types.QueryBalance}, req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().NoError(legacyAmino.UnmarshalJSON(res, &balance))
	suite.True(balance.IsEqual(newFooCoin(50)))
}

func (suite *KeeperTestSuite) TestQuerier_QueryAllBalances() {
	ctx, legacyAmino := suite.ctx, suite.encCfg.Amino
	_, _, addr := testdata.KeyTestPubAddr()
	req := abci.RequestQuery{
		Path: fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryAllBalances),
		Data: []byte{},
	}

	querier := keeper.NewQuerier(suite.bankKeeper, legacyAmino)

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

	suite.mockFundAccount(addr)
	suite.Require().NoError(testutil.FundAccount(suite.bankKeeper, ctx, addr, origCoins))
	res, err = querier(ctx, []string{types.QueryAllBalances}, req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().NoError(legacyAmino.UnmarshalJSON(res, &balances))
	suite.True(balances.IsEqual(origCoins))
}

func (suite *KeeperTestSuite) TestQuerier_QueryTotalSupply() {
	ctx, legacyAmino := suite.ctx, suite.encCfg.Amino

	genesisSupply, _, err := suite.bankKeeper.GetPaginatedTotalSupply(suite.ctx, &query.PageRequest{Limit: query.MaxLimit})
	suite.Require().NoError(err)

	testCoins := sdk.NewCoins(sdk.NewInt64Coin("test", 400000000))

	suite.mockMintCoins(mintAcc)
	suite.
		Require().
		NoError(suite.bankKeeper.MintCoins(ctx, minttypes.ModuleName, testCoins))

	req := abci.RequestQuery{
		Path: fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryTotalSupply),
		Data: []byte{},
	}

	querier := keeper.NewQuerier(suite.bankKeeper, legacyAmino)

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

func (suite *KeeperTestSuite) TestQuerier_QueryTotalSupplyOf() {
	ctx, legacyAmino := suite.ctx, suite.encCfg.Amino
	test1Supply := sdk.NewInt64Coin("test1", 4000000)
	test2Supply := sdk.NewInt64Coin("test2", 700000000)
	expectedTotalSupply := sdk.NewCoins(test1Supply, test2Supply)

	suite.mockMintCoins(mintAcc)
	suite.
		Require().
		NoError(suite.bankKeeper.MintCoins(ctx, minttypes.ModuleName, expectedTotalSupply))

	req := abci.RequestQuery{
		Path: fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QuerySupplyOf),
		Data: []byte{},
	}

	querier := keeper.NewQuerier(suite.bankKeeper, legacyAmino)

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

func (suite *KeeperTestSuite) TestQuerierRouteNotFound() {
	ctx := suite.ctx
	legacyAmino := suite.encCfg.Amino
	req := abci.RequestQuery{
		Path: fmt.Sprintf("custom/%s/invalid", types.ModuleName),
		Data: []byte{},
	}

	querier := keeper.NewQuerier(suite.bankKeeper, legacyAmino)
	_, err := querier(ctx, []string{"invalid"}, req)
	suite.Error(err)
}
