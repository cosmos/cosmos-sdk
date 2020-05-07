package client_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
)

type ClientTestSuite struct {
	suite.Suite

	cdc *codec.Codec
	ctx sdk.Context
	app *simapp.SimApp
}

func (suite *ClientTestSuite) SetupTest() {
	isCheckTx := false

	suite.app = simapp.Setup(isCheckTx)
	suite.cdc = suite.app.Codec()
	suite.ctx = suite.app.BaseApp.NewContext(isCheckTx, abci.Header{Height: 1, ChainID: "localhost_chain"})

}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

func (suite *ClientTestSuite) TestBeginBlocker() {
	localHostClient, found := suite.app.IBCKeeper.ClientKeeper.GetClientState(suite.ctx, exported.ClientTypeLocalHost)
	suite.Require().True(found)

	var prevHeight uint64
	for i := 0; i < 10; i++ {
		prevHeight = localHostClient.GetLatestHeight()
		_ = suite.app.Commit()
		res := suite.app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: suite.ctx.BlockHeight() + 1}})
		suite.Require().NotNil(res)

		suite.ctx = suite.ctx.WithBlockHeight(suite.app.LastBlockHeight() + 1)

		localHostClient, found = suite.app.IBCKeeper.ClientKeeper.GetClientState(suite.ctx, localHostClient.GetID())
		suite.Require().True(found)
		suite.Require().Equal(prevHeight+1, localHostClient.GetLatestHeight())
	}
}
