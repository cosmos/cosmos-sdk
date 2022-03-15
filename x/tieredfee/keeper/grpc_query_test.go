package keeper_test

import (
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/tieredfee/types"
)

func (suite *IntegrationTestSuite) TestQueryGenesisState() {
	suite.SetupTest() // reset
	var initialGasPrices []*sdk.ProtoDecCoins
	{
		rsp, err := suite.queryClient.Params(suite.ctx, &types.QueryParamsRequest{})
		suite.Require().NoError(err)
		suite.Require().Equal(types.DefaultParams(), *rsp.Params)

		for _, tier := range rsp.Params.Tiers {
			initialGasPrices = append(initialGasPrices, &sdk.ProtoDecCoins{Coins: tier.InitialGasPrice})
		}
	}

	{
		_, err := suite.queryClient.BlockGasUsed(suite.ctx, &types.QueryBlockGasUsedRequest{})
		// block gas used is not set yet
		suite.Require().Error(err)
	}

	{
		rsp, err := suite.queryClient.GasPrices(suite.ctx, &types.QueryGasPricesRequest{})
		suite.Require().NoError(err)
		// BeginBlock has been called in setup, so gas prices are initialized.
		suite.Require().Equal(initialGasPrices, rsp.GasPrices)
	}

	suite.app.EndBlock(abci.RequestEndBlock{})
	suite.app.Commit()
	suite.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{
		Height:  suite.app.LastBlockHeight() + 1,
		AppHash: suite.app.LastCommitID().Hash,
	}})

	{
		rsp, err := suite.queryClient.BlockGasUsed(suite.ctx, &types.QueryBlockGasUsedRequest{})
		suite.Require().NoError(err)
		suite.Require().Equal(uint64(0), rsp.GasUsed)
	}
}
