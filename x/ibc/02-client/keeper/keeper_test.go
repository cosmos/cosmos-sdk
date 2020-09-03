package keeper_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/keeper"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	localhosttypes "github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	testChainID = "gaiahub"

	testClientID  = "gaiachain"
	testClientID2 = "ethbridge"
	testClientID3 = "ethermint"

	height = 5

	trustingPeriod time.Duration = time.Hour * 24 * 7 * 2
	ubdPeriod      time.Duration = time.Hour * 24 * 7 * 3
	maxClockDrift  time.Duration = time.Second * 10
)

var testClientHeight = types.NewHeight(0, 5)

type KeeperTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain

	cdc            codec.Marshaler
	ctx            sdk.Context
	keeper         *keeper.Keeper
	consensusState *ibctmtypes.ConsensusState
	header         *ibctmtypes.Header
	valSet         *tmtypes.ValidatorSet
	valSetHash     tmbytes.HexBytes
	privVal        tmtypes.PrivValidator
	now            time.Time
	past           time.Time

	queryClient types.QueryClient
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)

	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(0))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(1))

	isCheckTx := false
	suite.now = time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)
	suite.past = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	now2 := suite.now.Add(time.Hour)
	app := simapp.Setup(isCheckTx)

	suite.cdc = app.AppCodec()
	suite.ctx = app.BaseApp.NewContext(isCheckTx, tmproto.Header{Height: height, ChainID: testClientID, Time: now2})
	suite.keeper = &app.IBCKeeper.ClientKeeper
	suite.privVal = tmtypes.NewMockPV()

	pubKey, err := suite.privVal.GetPubKey()
	suite.Require().NoError(err)

	validator := tmtypes.NewValidator(pubKey, 1)
	suite.valSet = tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})
	suite.valSetHash = suite.valSet.Hash()
	suite.header = ibctmtypes.CreateTestHeader(testChainID, height, height-1, now2, suite.valSet, suite.valSet, []tmtypes.PrivValidator{suite.privVal})
	suite.consensusState = ibctmtypes.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot([]byte("hash")), testClientHeight, suite.valSetHash)

	var validators stakingtypes.Validators
	for i := 1; i < 11; i++ {
		privVal := tmtypes.NewMockPV()
		pk, err := privVal.GetPubKey()
		suite.Require().NoError(err)
		val := stakingtypes.NewValidator(sdk.ValAddress(pk.Address()), pk, stakingtypes.Description{})
		val.Status = sdk.Bonded
		val.Tokens = sdk.NewInt(rand.Int63())
		validators = append(validators, val)

		app.StakingKeeper.SetHistoricalInfo(suite.ctx, int64(i), stakingtypes.NewHistoricalInfo(suite.ctx.BlockHeader(), validators))
	}

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.IBCKeeper.ClientKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestSetClientState() {
	clientState := ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, types.Height{}, commitmenttypes.GetSDKSpecs())
	suite.keeper.SetClientState(suite.ctx, testClientID, clientState)

	retrievedState, found := suite.keeper.GetClientState(suite.ctx, testClientID)
	suite.Require().True(found, "GetClientState failed")
	suite.Require().Equal(clientState, retrievedState, "Client states are not equal")
}

func (suite *KeeperTestSuite) TestSetClientType() {
	suite.keeper.SetClientType(suite.ctx, testClientID, exported.Tendermint)
	clientType, found := suite.keeper.GetClientType(suite.ctx, testClientID)

	suite.Require().True(found, "GetClientType failed")
	suite.Require().Equal(exported.Tendermint, clientType, "ClientTypes not stored correctly")
}

func (suite *KeeperTestSuite) TestSetClientConsensusState() {
	suite.keeper.SetClientConsensusState(suite.ctx, testClientID, height, suite.consensusState)

	retrievedConsState, found := suite.keeper.GetClientConsensusState(suite.ctx, testClientID, height)
	suite.Require().True(found, "GetConsensusState failed")

	tmConsState, ok := retrievedConsState.(*ibctmtypes.ConsensusState)
	suite.Require().True(ok)
	suite.Require().Equal(suite.consensusState, tmConsState, "ConsensusState not stored correctly")
}

func (suite *KeeperTestSuite) TestValidateSelfClient() {
	testCases := []struct {
		name        string
		clientState exported.ClientState
		expPass     bool
	}{
		{
			"success",
			ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, testClientHeight, commitmenttypes.GetSDKSpecs()),
			true,
		},
		{
			"invalid client type",
			localhosttypes.NewClientState(testChainID, testClientHeight),
			false,
		},
		{
			"frozen client",
			&ibctmtypes.ClientState{testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, testClientHeight, testClientHeight, commitmenttypes.GetSDKSpecs()},
			false,
		},
		{
			"incorrect chainID",
			ibctmtypes.NewClientState("gaiatestnet", ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, testClientHeight, commitmenttypes.GetSDKSpecs()),
			false,
		},
		{
			"invalid client height",
			ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, types.NewHeight(0, testClientHeight.EpochHeight+10), commitmenttypes.GetSDKSpecs()),
			false,
		},
		{
			"invalid proof specs",
			ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, testClientHeight, nil),
			false,
		},
		{
			"invalid trust level",
			ibctmtypes.NewClientState(testChainID, ibctmtypes.Fraction{0, 1}, trustingPeriod, ubdPeriod, maxClockDrift, testClientHeight, commitmenttypes.GetSDKSpecs()),
			false,
		},
		{
			"invalid unbonding period",
			ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod+10, maxClockDrift, testClientHeight, commitmenttypes.GetSDKSpecs()),
			false,
		},
		{
			"invalid trusting period",
			ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, ubdPeriod+10, ubdPeriod, maxClockDrift, testClientHeight, commitmenttypes.GetSDKSpecs()),
			false,
		},
	}

	ctx := suite.ctx.WithChainID(testChainID)
	ctx = ctx.WithBlockHeight(height)

	for _, tc := range testCases {
		err := suite.keeper.ValidateSelfClient(ctx, tc.clientState)
		if tc.expPass {
			suite.Require().NoError(err, "expected valid client for case: %s", tc.name)
		} else {
			suite.Require().Error(err, "expected invalid client for case: %s", tc.name)
		}
	}
}

func (suite KeeperTestSuite) TestGetAllClients() {
	clientIDs := []string{
		testClientID2, testClientID3, testClientID,
	}
	expClients := []exported.ClientState{
		ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, types.Height{}, commitmenttypes.GetSDKSpecs()),
		ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, types.Height{}, commitmenttypes.GetSDKSpecs()),
		ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, types.Height{}, commitmenttypes.GetSDKSpecs()),
	}

	for i := range expClients {
		suite.keeper.SetClientState(suite.ctx, clientIDs[i], expClients[i])
	}

	// add localhost client
	localHostClient, found := suite.keeper.GetClientState(suite.ctx, exported.ClientTypeLocalHost)
	suite.Require().True(found)
	expClients = append(expClients, localHostClient)

	clients := suite.keeper.GetAllClients(suite.ctx)
	suite.Require().Len(clients, len(expClients))
	suite.Require().Equal(expClients, clients)
}

func (suite KeeperTestSuite) TestGetAllGenesisClients() {
	clientIDs := []string{
		testClientID2, testClientID3, testClientID,
	}
	expClients := []exported.ClientState{
		ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, types.Height{}, commitmenttypes.GetSDKSpecs()),
		ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, types.Height{}, commitmenttypes.GetSDKSpecs()),
		ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, types.Height{}, commitmenttypes.GetSDKSpecs()),
	}

	expGenClients := make([]types.IdentifiedClientState, len(expClients))

	for i := range expClients {
		suite.keeper.SetClientState(suite.ctx, clientIDs[i], expClients[i])
		expGenClients[i] = types.NewIdentifiedClientState(clientIDs[i], expClients[i])
	}

	// add localhost client
	localHostClient, found := suite.keeper.GetClientState(suite.ctx, exported.ClientTypeLocalHost)
	suite.Require().True(found)
	expGenClients = append(expGenClients, types.NewIdentifiedClientState(exported.ClientTypeLocalHost, localHostClient))

	genClients := suite.keeper.GetAllGenesisClients(suite.ctx)

	suite.Require().Equal(expGenClients, genClients)
}

func (suite KeeperTestSuite) TestGetConsensusState() {
	suite.ctx = suite.ctx.WithBlockHeight(10)
	cases := []struct {
		name    string
		height  uint64
		expPass bool
	}{
		{"zero height", 0, false},
		{"height > latest height", uint64(suite.ctx.BlockHeight()) + 1, false},
		{"latest height - 1", uint64(suite.ctx.BlockHeight()) - 1, true},
		{"latest height", uint64(suite.ctx.BlockHeight()), true},
	}

	for i, tc := range cases {
		tc := tc
		cs, found := suite.keeper.GetSelfConsensusState(suite.ctx, tc.height)
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
	clientState := ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, testClientHeight, commitmenttypes.GetSDKSpecs())

	suite.keeper.SetClientState(suite.ctx, testClientID, clientState)
	suite.keeper.SetClientConsensusState(suite.ctx, testClientID, height, suite.consensusState)

	nextState := ibctmtypes.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot([]byte("next")), types.NewHeight(0, height+5), suite.valSetHash)

	header := ibctmtypes.CreateTestHeader(testClientID, height+5, height, suite.header.Header.Time.Add(time.Minute),
		suite.valSet, suite.valSet, []tmtypes.PrivValidator{suite.privVal})

	// mock update functionality
	clientState.LatestHeight = types.NewHeight(0, header.GetHeight())
	suite.keeper.SetClientConsensusState(suite.ctx, testClientID, height+5, nextState)
	suite.keeper.SetClientState(suite.ctx, testClientID, clientState)

	latest, ok := suite.keeper.GetLatestClientConsensusState(suite.ctx, testClientID)
	suite.Require().True(ok)
	suite.Require().Equal(nextState, latest, "Latest client not returned correctly")

	// Should return existing consensusState at latestClientHeight
	lte, ok := suite.keeper.GetClientConsensusStateLTE(suite.ctx, testClientID, height+3)
	suite.Require().True(ok)
	suite.Require().Equal(suite.consensusState, lte, "LTE helper function did not return latest client state below height: %d", height+3)
}

func (suite KeeperTestSuite) TestGetAllConsensusStates() {
	expConsensus := []exported.ConsensusState{
		ibctmtypes.NewConsensusState(
			suite.consensusState.Timestamp, commitmenttypes.NewMerkleRoot([]byte("hash")), suite.consensusState.Height, nil,
		),
		ibctmtypes.NewConsensusState(
			suite.consensusState.Timestamp.Add(time.Minute), commitmenttypes.NewMerkleRoot([]byte("app_hash")), suite.consensusState.Height.Increment(), nil,
		),
	}

	expConsensus2 := []exported.ConsensusState{
		ibctmtypes.NewConsensusState(
			suite.consensusState.Timestamp.Add(2*time.Minute), commitmenttypes.NewMerkleRoot([]byte("app_hash_2")), types.NewHeight(0, suite.consensusState.GetHeight()+2), nil,
		),
	}

	expAnyConsensus := types.ClientsConsensusStates{
		types.NewClientConsensusStates(testClientID, expConsensus),
		types.NewClientConsensusStates(testClientID2, expConsensus2),
	}.Sort()

	suite.keeper.SetClientConsensusState(suite.ctx, testClientID, expConsensus[0].GetHeight(), expConsensus[0])
	suite.keeper.SetClientConsensusState(suite.ctx, testClientID, expConsensus[1].GetHeight(), expConsensus[1])
	suite.keeper.SetClientConsensusState(suite.ctx, testClientID2, expConsensus2[0].GetHeight(), expConsensus2[0])

	consStates := suite.keeper.GetAllConsensusStates(suite.ctx)
	suite.Require().Equal(expAnyConsensus, consStates, "%s \n\n%s", expAnyConsensus, consStates)
}
