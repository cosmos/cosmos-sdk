package client_test

import (
	"testing"

	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ClientTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain
}

func (suite *ClientTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)

	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(0))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(1))
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

/* TODO: uncomment once simapp is switched to proto
func (suite *ClientTestSuite) TestBeginBlocker() {
	prevHeight := suite.ctx.BlockHeight()

	localHostClient, found := suite.app.IBCKeeper.ClientKeeper.GetClientState(suite.ctx, exported.ClientTypeLocalHost)
	suite.Require().True(found)
	suite.Require().Equal(prevHeight, int64(localHostClient.GetLatestHeight()))

	for i := 0; i < 10; i++ {
		// increment height
		suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)

		suite.Require().NotPanics(func() {
			client.BeginBlocker(suite.ctx, suite.app.IBCKeeper.ClientKeeper)
		}, "BeginBlocker shouldn't panic")

		localHostClient, found = suite.app.IBCKeeper.ClientKeeper.GetClientState(suite.ctx, exported.ClientTypeLocalHost)
		suite.Require().True(found)
		suite.Require().Equal(prevHeight+1, int64(localHostClient.GetLatestHeight()))
		prevHeight = int64(localHostClient.GetLatestHeight())
	}
}
*/
