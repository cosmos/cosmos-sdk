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
	localhosttypes "github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/types"
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
	localHostClient := localhosttypes.NewClientState(
		suite.app.IBCKeeper.ClientKeeper.ClientStore(suite.ctx, exported.ClientTypeLocalHost),
		suite.ctx.ChainID(),
		suite.ctx.BlockHeight(),
	)
	_, err := suite.app.IBCKeeper.ClientKeeper.CreateClient(suite.ctx, localHostClient, nil)
	suite.Require().NoError(err)

	// increase height
	suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)

	var prevHeight uint64
	for i := 0; i < 10; i++ {
		prevHeight = localHostClient.GetLatestHeight()
		suite.Require().NotPanics(func() {
			client.BeginBlocker(suite.ctx, suite.app.IBCKeeper.ClientKeeper)
		}, "BeginBlocker shouldn't panic")
		localHostClient, found := suite.app.IBCKeeper.ClientKeeper.GetClientState(suite.ctx, localHostClient.GetID())
		suite.Require().True(found)
		suite.Require().Equal(prevHeight+1, localHostClient.GetLatestHeight())
	}
}
