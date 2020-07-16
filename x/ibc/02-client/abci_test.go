package client_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
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
	suite.ctx = suite.app.BaseApp.NewContext(isCheckTx, abci.Header{Height: 0, ChainID: "localhost_chain"})

}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

func (suite *ClientTestSuite) TestBeginBlocker() {
	prevHeight := suite.ctx.BlockHeight()

	localHostClient, found := suite.app.IBCKeeper.ClientKeeper.GetClientState(suite.ctx, exported.ClientTypeLocalHost)
	suite.Require().True(found)
	suite.Require().Equal(exported.NewHeight(0, uint64(prevHeight)), localHostClient.GetLatestHeight())

	for i := 0; i < 10; i++ {
		// increment height
		suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)

		suite.Require().NotPanics(func() {
			client.BeginBlocker(suite.ctx, suite.app.IBCKeeper.ClientKeeper)
		}, "BeginBlocker shouldn't panic")

		localHostClient, found = suite.app.IBCKeeper.ClientKeeper.GetClientState(suite.ctx, exported.ClientTypeLocalHost)
		suite.Require().True(found)
		suite.Require().Equal(exported.NewHeight(0, uint64(prevHeight+1)), localHostClient.GetLatestHeight())
		prevHeight = int64(localHostClient.GetLatestHeight().EpochHeight)
	}
}
