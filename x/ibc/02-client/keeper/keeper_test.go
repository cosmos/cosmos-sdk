package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	localhosttypes "github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
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

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestSetClientState() {
	clientState := ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, types.ZeroHeight(), commitmenttypes.GetSDKSpecs(), false, false)
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
	consensusState := suite.chainA.ConsensusStateFromCurrentHeader()
	suite.chainA.App.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], ibctesting.ClientHeight, consensusState)

	retrievedConsState, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientConsensusState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], ibctesting.ClientHeight)
	suite.Require().True(found, "GetConsensusState failed")

	tmConsState, ok := retrievedConsState.(*ibctmtypes.ConsensusState)
	suite.Require().True(ok)
	suite.Require().Equal(consensusState, tmConsState, "ConsensusState not stored correctly")
}

func (suite *KeeperTestSuite) TestValidateSelfClient() {
	testCases := []struct {
		name        string
		clientState exported.ClientState
		expPass     bool
	}{
		{
			"success",
			ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, ibctesting.ClientHeight, commitmenttypes.GetSDKSpecs(), false, false),
			true,
		},
		{
			"invalid client type",
			localhosttypes.NewClientState(suite.chainA.ChainID, ibctesting.ClientHeight),
			false,
		},
		{
			"frozen client",
			&ibctmtypes.ClientState{suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, ibctesting.ClientHeight, ibctesting.ClientHeight, commitmenttypes.GetSDKSpecs(), false, false},
			false,
		},
		{
			"incorrect chainID",
			ibctmtypes.NewClientState("gaiatestnet", ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, ibctesting.ClientHeight, commitmenttypes.GetSDKSpecs(), false, false),
			false,
		},
		{
			"invalid client ibctesting.Height",
			ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, types.NewHeight(0, ibctesting.ClientHeight.EpochHeight+10), commitmenttypes.GetSDKSpecs(), false, false),
			false,
		},
		{
			"invalid proof specs",
			ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, ibctesting.ClientHeight, nil, false, false),
			false,
		},
		{
			"invalid trust level",
			ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.Fraction{0, 1}, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, ibctesting.ClientHeight, commitmenttypes.GetSDKSpecs(), false, false),
			false,
		},
		{
			"invalid unbonding period",
			ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod+10, ibctesting.MaxClockDrift, ibctesting.ClientHeight, commitmenttypes.GetSDKSpecs(), false, false),
			false,
		},
		{
			"invalid trusting period",
			ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.UnbondingPeriod+10, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, ibctesting.ClientHeight, commitmenttypes.GetSDKSpecs(), false, false),
			false,
		},
	}

	ctx := suite.chainA.GetContext().WithChainID(suite.chainA.ChainID)
	ctx = ctx.WithBlockHeight(ibctesting.Height)

	for _, tc := range testCases {
		err := suite.chainA.App.IBCKeeper.ClientKeeper.ValidateSelfClient(ctx, tc.clientState)
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
		ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, types.ZeroHeight(), commitmenttypes.GetSDKSpecs(), false, false),
		ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, types.ZeroHeight(), commitmenttypes.GetSDKSpecs(), false, false),
		ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, types.ZeroHeight(), commitmenttypes.GetSDKSpecs(), false, false),
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
	clientIDs := []string{"client-1", "client-2", "client-3"}
	expClients := []exported.ClientState{
		ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, types.ZeroHeight(), commitmenttypes.GetSDKSpecs(), false, false),
		ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, types.ZeroHeight(), commitmenttypes.GetSDKSpecs(), false, false),
		ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, types.ZeroHeight(), commitmenttypes.GetSDKSpecs(), false, false),
	}

	expGenClients := make([]types.IdentifiedClientState, len(expClients))

	for i := range expClients {
		suite.chainA.App.IBCKeeper.ClientKeeper.SetClientState(suite.chainA.GetContext(), clientIDs[i], expClients[i])
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
		height  types.Height
		expPass bool
	}{
		{"zero ibctesting.Height", types.ZeroHeight(), false},
		{"ibctesting.Height > latest ibctesting.Height", types.NewHeight(0, uint64(suite.chainA.GetContext().BlockHeight())+1), false},
		{"latest ibctesting.Height - 1", types.NewHeight(0, uint64(suite.chainA.GetContext().BlockHeight())-1), true},
		{"latest ibctesting.Height", types.GetSelfHeight(suite.chainA.GetContext()), true},
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
	clientA, _ := suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)

	// initial setup
	clientState := ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, ibctesting.ClientHeight, commitmenttypes.GetSDKSpecs(), false, false)

	consensusState := suite.chainA.ConsensusStateFromCurrentHeader()
	suite.chainA.App.IBCKeeper.ClientKeeper.SetClientState(suite.chainA.GetContext(), clientA, clientState)
	suite.chainA.App.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.chainA.GetContext(), clientA, ibctesting.ClientHeight, consensusState)

	nextState := ibctmtypes.NewConsensusState(suite.chainA.CurrentHeader.Time, consensusState.GetRoot().(commitmenttypes.MerkleRoot), ibctesting.ClientHeight, suite.chainA.Vals.Hash())

	header := suite.chainA.CreateTMClientHeader()

	// mock update functionality
	clientState.LatestHeight = header.GetHeight().(types.Height)
	suite.chainA.App.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.chainA.GetContext(), clientA, header.GetHeight(), nextState)
	suite.chainA.App.IBCKeeper.ClientKeeper.SetClientState(suite.chainA.GetContext(), clientA, clientState)

	latest, ok := suite.chainA.App.IBCKeeper.ClientKeeper.GetLatestClientConsensusState(suite.chainA.GetContext(), clientA)
	suite.Require().True(ok)
	suite.Require().Equal(nextState, latest, "Latest client not returned correctly")

	// Should return existing consensusState at latestClientHeight
	lte, ok := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientConsensusStateLTE(suite.chainA.GetContext(), clientA, types.NewHeight(0, ibctesting.Height+3))
	suite.Require().True(ok)
	suite.Require().Equal(consensusState, lte, "LTE helper function did not return latest client state below ibctesting.Height: %d", ibctesting.Height+3)
}

func (suite KeeperTestSuite) TestGetAllConsensusStates() {
	clientA, _ := suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)

	consensusState := suite.chainA.ConsensusStateFromCurrentHeader()
	clientHeight, ok := consensusState.GetHeight().(types.Height)
	suite.Require().True(ok)

	timestamp := time.Unix(0, int64(consensusState.GetTimestamp()))

	expConsensus := []exported.ConsensusState{
		ibctmtypes.NewConsensusState(
			timestamp, commitmenttypes.NewMerkleRoot([]byte("hash")), clientHeight, nil,
		),
		ibctmtypes.NewConsensusState(
			timestamp.Add(time.Minute), commitmenttypes.NewMerkleRoot([]byte("app_hash")), clientHeight.Increment(), nil,
		),
	}

	expConsensus2 := []exported.ConsensusState{
		ibctmtypes.NewConsensusState(
			timestamp.Add(2*time.Minute), commitmenttypes.NewMerkleRoot([]byte("app_hash_2")), types.NewHeight(0, clientHeight.GetEpochHeight()+2), nil,
		),
	}

	expAnyConsensus := types.ClientsConsensusStates{
		types.NewClientConsensusStates(clientA, expConsensus),
		types.NewClientConsensusStates(clientA, expConsensus2),
	}.Sort()

	suite.chainA.App.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.chainA.GetContext(), clientA, expConsensus[0].GetHeight(), expConsensus[0])
	suite.chainA.App.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.chainA.GetContext(), clientA, expConsensus[1].GetHeight(), expConsensus[1])
	suite.chainA.App.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.chainA.GetContext(), clientA, expConsensus2[0].GetHeight(), expConsensus2[0])

	consStates := suite.chainA.App.IBCKeeper.ClientKeeper.GetAllConsensusStates(suite.chainA.GetContext())
	suite.Require().Equal(expAnyConsensus.Len(), consStates.Len())
	suite.Require().Equal(expAnyConsensus, consStates, "%s \n\n%s", expAnyConsensus, consStates)
}
