package client_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	client "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
	localhosttypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/09-localhost/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
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

	// set localhost client
	revision := types.ParseChainID(suite.chainA.GetContext().ChainID())
	localHostClient := localhosttypes.NewClientState(
		suite.chainA.GetContext().ChainID(), types.NewHeight(revision, uint64(suite.chainA.GetContext().BlockHeight())),
	)
	suite.chainA.App.IBCKeeper.ClientKeeper.SetClientState(suite.chainA.GetContext(), exported.Localhost, localHostClient)
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

func (suite *ClientTestSuite) TestBeginBlocker() {
	prevHeight := types.GetSelfHeight(suite.chainA.GetContext())

	localHostClient := suite.chainA.GetClientState(exported.Localhost)
	suite.Require().Equal(prevHeight, localHostClient.GetLatestHeight())

	for i := 0; i < 10; i++ {
		// increment height
		suite.coordinator.CommitBlock(suite.chainA, suite.chainB)

		suite.Require().NotPanics(func() {
			client.BeginBlocker(suite.chainA.GetContext(), suite.chainA.App.IBCKeeper.ClientKeeper)
		}, "BeginBlocker shouldn't panic")

		localHostClient = suite.chainA.GetClientState(exported.Localhost)
		suite.Require().Equal(prevHeight.Increment(), localHostClient.GetLatestHeight())
		prevHeight = localHostClient.GetLatestHeight().(types.Height)
	}
}

func (suite *ClientTestSuite) TestBeginBlockerConsensusState() {
	plan := &upgradetypes.Plan{
		Name:   "test",
		Height: suite.chainA.GetContext().BlockHeight() + 1,
	}
	store := suite.chainA.GetContext().KVStore(suite.chainA.App.GetKey(upgradetypes.StoreKey))
	bz := suite.chainA.App.AppCodec().MustMarshalBinaryBare(plan)
	store.Set(upgradetypes.PlanKey(), bz)

	nextValsHash := []byte("nextValsHash")
	newCtx := suite.chainA.GetContext().WithBlockHeader(tmproto.Header{
		Height:             suite.chainA.GetContext().BlockHeight(),
		NextValidatorsHash: nextValsHash,
	})

	err := suite.chainA.App.UpgradeKeeper.SetUpgradedClient(newCtx, plan.Height, []byte("client state"))
	suite.Require().NoError(err)

	req := abci.RequestBeginBlock{Header: newCtx.BlockHeader()}
	suite.chainA.App.BeginBlock(req)

	// plan Height is at ctx.BlockHeight+1
	consState, found := suite.chainA.App.UpgradeKeeper.GetUpgradedConsensusState(newCtx, plan.Height)
	suite.Require().True(found)
	bz, err = types.MarshalConsensusState(suite.chainA.App.AppCodec(), &ibctmtypes.ConsensusState{Timestamp: newCtx.BlockTime(), NextValidatorsHash: nextValsHash})
	suite.Require().NoError(err)
	suite.Require().Equal(bz, consState)
}
