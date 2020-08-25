package keeper_test

import (
	"time"

	"github.com/stretchr/testify/suite"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	localhosttypes "github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

type KeeperTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(0))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(1))
}

func (suite *KeeperTestSuite) TestSetClientState() {
	clientState := ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, 0, commitmenttypes.GetSDKSpecs())
	suite.chainA.App.IBCKeeper.ClientKeeper.SetClientState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState)

	retrievedState, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0])
	suite.Require().True(found, "GetClientState failed")
	suite.Require().Equal(clientState, retrievedState, "Client states are not equal")
}

func (suite *KeeperTestSuite) TestSetClientType() {
	suite.chainA.App.IBCKeeper.ClientKeeper.SetClientType(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], exported.Tendermint)
	clientType, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientType(suite.chainA.GetContext(), suite.chainA.ClientIDs[0])

	suite.Require().True(found, "GetClientType failed")
	suite.Require().Equal(exported.Tendermint, clientType, "ClientTypes not stored correctly")
}

func (suite *KeeperTestSuite) TestSetClientConsensusState() {
	suite.chainA.App.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], testClientHeight, suite.consensusState)

	retrievedConsState, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientConsensusState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], testClientHeight)
	suite.Require().True(found, "GetConsensusState failed")

	tmConsState, ok := retrievedConsState.(*ibctmtypes.ConsensusState)
	suite.Require().True(ok)
	suite.Require().Equal(suite.consensusState, tmConsState, "ConsensusState not stored correctly")
}

func (suite *KeeperTestSuite) TestValidateSelfClient() {
	testCases := []struct {
		name        string
		clientState clientexported.ClientState
		expPass     bool
	}{
		{
			"success",
			ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, uint64(testClientHeight), commitmenttypes.GetSDKSpecs()),
			true,
		},
		{
			"invalid client type",
			localhosttypes.NewClientState(testChainID, testClientHeight),
			false,
		},
		{
			"frozen client",
			&ibctmtypes.ClientState{testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, uint64(testClientHeight), uint64(testClientHeight), commitmenttypes.GetSDKSpecs()},
			false,
		},
		{
			"incorrect chainID",
			ibctmtypes.NewClientState("gaiatestnet", ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, uint64(testClientHeight), commitmenttypes.GetSDKSpecs()),
			false,
		},
		{
			"invalid client height",
			ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, uint64(testClientHeight)+10, commitmenttypes.GetSDKSpecs()),
			false,
		},
		{
			"invalid proof specs",
			ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, uint64(testClientHeight), nil),
			false,
		},
		{
			"invalid trust level",
			ibctmtypes.NewClientState(testChainID, ibctmtypes.Fraction{0, 1}, trustingPeriod, ubdPeriod, maxClockDrift, uint64(testClientHeight), commitmenttypes.GetSDKSpecs()),
			false,
		},
		{
			"invalid unbonding period",
			ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod+10, maxClockDrift, uint64(testClientHeight), commitmenttypes.GetSDKSpecs()),
			false,
		},
		{
			"invalid trusting period",
			ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, ubdPeriod+10, ubdPeriod, maxClockDrift, uint64(testClientHeight), commitmenttypes.GetSDKSpecs()),
			false,
		},
	}

	ctx := suite.chainA.GetContext().WithChainID(testChainID)
	ctx = ctx.WithBlockHeight(testClientHeight)

	for _, tc := range testCases {
		err := suite.chainA.App.IBCKeeper.ValidateSelfClient(ctx, tc.clientState)
		if tc.expPass {
			suite.Require().NoError(err, "expected valid client for case: %s", tc.name)
		} else {
			suite.Require().Error(err, "expected invalid client for case: %s", tc.name)
		}
	}
}

func (suite KeeperTestSuite) TestGetAllClients() {
	clientIDs := []string{
		suite.chainA.ClientIDs[0], suite.chainA.ClientIDs[0], suite.chainA.ClientIDs[0],
	}
	expClients := []exported.ClientState{
		ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, 0, commitmenttypes.GetSDKSpecs()),
		ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, 0, commitmenttypes.GetSDKSpecs()),
		ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, 0, commitmenttypes.GetSDKSpecs()),
	}

	for i := range expClients {
		suite.chainA.App.IBCKeeper.ClientKeeper.SetClientState(suite.chainA.GetContext(), clientIDs[i], expClients[i])
	}

	// add localhost client
	localHostClient, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientState(suite.chainA.GetContext(), exported.ClientTypeLocalHost)
	suite.Require().True(found)
	expClients = append(expClients, localHostClient)

	clients := suite.chainA.App.IBCKeeper.ClientKeeper.GetAllClients(suite.chainA.GetContext())
	suite.Require().Len(clients, len(expClients))
	suite.Require().Equal(expClients, clients)
}

func (suite KeeperTestSuite) TestGetAllGenesisClients() {
	clientIDs := []string{
		suite.chainA.ClientIDs[0], suite.chainA.ClientIDs[0], suite.chainA.ClientIDs[0],
	}
	expClients := []exported.ClientState{
		ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, 0, commitmenttypes.GetSDKSpecs()),
		ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, 0, commitmenttypes.GetSDKSpecs()),
		ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, 0, commitmenttypes.GetSDKSpecs()),
	}

	expGenClients := make([]types.IdentifiedClientState, len(expClients))

	for i := range expClients {
		suite.chainA.App.IBCKeeper.SetClientState(suite.chainA.GetContext(), clientIDs[i], expClients[i])
		expGenClients[i] = types.NewIdentifiedClientState(clientIDs[i], expClients[i])
	}

	// add localhost client
	localHostClient, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientState(suite.chainA.GetContext(), exported.ClientTypeLocalHost)
	suite.Require().True(found)
	expGenClients = append(expGenClients, types.NewIdentifiedClientState(exported.ClientTypeLocalHost, localHostClient))

	genClients := suite.chainA.App.IBCKeeper.ClientKeeper.GetAllGenesisClients(suite.chainA.GetContext())

	suite.Require().Equal(expGenClients, genClients)
}

func (suite KeeperTestSuite) TestGetConsensusState() {
	ctx := suite.chainA.GetContext().WithBlockHeight(10)
	cases := []struct {
		name    string
		height  uint64
		expPass bool
	}{
		{"zero height", 0, false},
		{"height > latest height", uint64(ctx.BlockHeight()) + 1, false},
		{"latest height - 1", uint64(ctx.BlockHeight()) - 1, true},
		{"latest height", uint64(ctx.BlockHeight()), true},
	}

	for i, tc := range cases {
		tc := tc
		cs, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetSelfConsensusState(ctx, tc.height)
		if tc.expPass {
			suite.Require().True(found, "Case %d should have passed: %s", i, tc.name)
			suite.Require().NotNil(cs, "Case %d should have passed: %s", i, tc.name)
		} else {
			suite.Require().False(found, "Case %d should have failed: %s", i, tc.name)
			suite.Require().Nil(cs, "Case %d should have failed: %s", i, tc.name)
		}
	}
}

func (suite KeeperTestSuite) TestConsensusStateHelpers() {
	// initial setup
	clientState := ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, testClientHeight, commitmenttypes.GetSDKSpecs())

	suite.chainA.App.IBCKeeper.ClientKeeper.SetClientState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState)
	suite.chainA.App.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], testClientHeight, suite.consensusState)

	nextState := ibctmtypes.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot([]byte("next")), testClientHeight+5, suite.valSetHash)

	header := ibctmtypes.CreateTestHeader(suite.chainA.ClientIDs[0], testClientHeight+5, testClientHeight, suite.header.GetTime().Add(time.Minute),
		suite.valSet, suite.valSet, []tmtypes.PrivValidator{suite.privVal})

	// mock update functionality
	clientState.LatestHeight = header.GetHeight()
	suite.chainA.App.IBCKeeper.SetClientConsensusState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], testClientHeight+5, nextState)
	suite.chainA.App.IBCKeeper.SetClientState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState)

	latest, ok := suite.chainA.App.IBCKeeper.ClientKeeper.GetLatestClientConsensusState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0])
	suite.Require().True(ok)
	suite.Require().Equal(nextState, latest, "Latest client not returned correctly")

	// Should return existing consensusState at latestClientHeight
	lte, ok := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientConsensusStateLTE(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], testClientHeight+3)
	suite.Require().True(ok)
	suite.Require().Equal(suite.consensusState, lte, "LTE helper function did not return latest client state below height: %d", testClientHeight+3)
}

func (suite KeeperTestSuite) TestGetAllConsensusStates() {
	expConsensus := []exported.ConsensusState{
		ibctmtypes.NewConsensusState(
			suite.consensusState.Timestamp, commitmenttypes.NewMerkleRoot([]byte("hash")), suite.consensusState.GetHeight(), nil,
		),
		ibctmtypes.NewConsensusState(
			suite.consensusState.Timestamp.Add(time.Minute), commitmenttypes.NewMerkleRoot([]byte("app_hash")), suite.consensusState.GetHeight()+1, nil,
		),
	}

	expConsensus2 := []exported.ConsensusState{
		ibctmtypes.NewConsensusState(
			suite.consensusState.Timestamp.Add(2*time.Minute), commitmenttypes.NewMerkleRoot([]byte("app_hash_2")), suite.consensusState.GetHeight()+2, nil,
		),
	}

	expAnyConsensus := types.ClientsConsensusStates{
		types.NewClientConsensusStates(suite.chainA.ClientIDs[0], expConsensus),
		types.NewClientConsensusStates(suite.chainA.ClientIDs[0], expConsensus2),
	}.Sort()

	suite.chainA.App.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], expConsensus[0].GetHeight(), expConsensus[0])
	suite.chainA.App.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], expConsensus[1].GetHeight(), expConsensus[1])
	suite.chainA.App.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], expConsensus2[0].GetHeight(), expConsensus2[0])

	consStates := suite.chainA.App.IBCKeeper.ClientKeeper.GetAllConsensusStates(suite.chainA.GetContext())
	suite.Require().Equal(expAnyConsensus, consStates, "%s \n\n%s", expAnyConsensus, consStates)
}
