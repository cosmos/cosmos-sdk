package keeper_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/keeper"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	testChainID = "gaiahub"

	testClientID  = "gaiachain"
	testClientID2 = "ethbridge"
	testClientID3 = "ethermint"

	testClientHeight = 5

	trustingPeriod time.Duration = time.Hour * 24 * 7 * 2
	ubdPeriod      time.Duration = time.Hour * 24 * 7 * 3
	maxClockDrift  time.Duration = time.Second * 10
)

type KeeperTestSuite struct {
	suite.Suite

	cdc            codec.Marshaler
	aminoCdc       *codec.Codec
	ctx            sdk.Context
	keeper         *keeper.Keeper
	consensusState ibctmtypes.ConsensusState
	header         ibctmtypes.Header
	valSet         *tmtypes.ValidatorSet
	valSetHash     tmbytes.HexBytes
	privVal        tmtypes.PrivValidator
	now            time.Time
	past           time.Time
}

func (suite *KeeperTestSuite) SetupTest() {
	isCheckTx := false
	suite.now = time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)
	suite.past = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	now2 := suite.now.Add(time.Hour)
	app := simapp.Setup(isCheckTx)

	suite.cdc = app.AppCodec()
	suite.aminoCdc = app.Codec()
	suite.ctx = app.BaseApp.NewContext(isCheckTx, abci.Header{Height: testClientHeight, ChainID: testClientID, Time: now2})
	suite.keeper = &app.IBCKeeper.ClientKeeper
	suite.privVal = tmtypes.NewMockPV()

	pubKey, err := suite.privVal.GetPubKey()
	suite.Require().NoError(err)

	validator := tmtypes.NewValidator(pubKey, 1)
	suite.valSet = tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})
	suite.valSetHash = suite.valSetHash
	suite.header = ibctmtypes.CreateTestHeader(testChainID, testClientHeight, testClientHeight-1, now2, suite.valSet, suite.valSet, []tmtypes.PrivValidator{suite.privVal})
	suite.consensusState = ibctmtypes.ConsensusState{
		Height:             testClientHeight,
		Timestamp:          suite.now,
		Root:               commitmenttypes.NewMerkleRoot([]byte("hash")),
		NextValidatorsHash: suite.valSetHash,
	}

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
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestSetClientState() {
	clientState := ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, 0, commitmenttypes.GetSDKSpecs())
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
	suite.keeper.SetClientConsensusState(suite.ctx, testClientID, testClientHeight, suite.consensusState)

	retrievedConsState, found := suite.keeper.GetClientConsensusState(suite.ctx, testClientID, testClientHeight)
	suite.Require().True(found, "GetConsensusState failed")

	tmConsState, ok := retrievedConsState.(ibctmtypes.ConsensusState)
	suite.Require().True(ok)
	suite.Require().Equal(suite.consensusState, tmConsState, "ConsensusState not stored correctly")
}

func (suite KeeperTestSuite) TestGetAllClients() {
	clientIDs := []string{
		testClientID2, testClientID3, testClientID,
	}
	expClients := []exported.ClientState{
		ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, 0, commitmenttypes.GetSDKSpecs()),
		ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, 0, commitmenttypes.GetSDKSpecs()),
		ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, 0, commitmenttypes.GetSDKSpecs()),
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
		ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, 0, commitmenttypes.GetSDKSpecs()),
		ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, 0, commitmenttypes.GetSDKSpecs()),
		ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, 0, commitmenttypes.GetSDKSpecs()),
	}

	expGenClients := make([]types.GenesisClientState, len(expClients))

	for i := range expClients {
		suite.keeper.SetClientState(suite.ctx, clientIDs[i], expClients[i])
		expGenClients[i] = types.NewGenesisClientState(clientIDs[i], expClients[i])
	}

	// add localhost client
	localHostClient, found := suite.keeper.GetClientState(suite.ctx, exported.ClientTypeLocalHost)
	suite.Require().True(found)
	expGenClients = append(expGenClients, types.NewGenesisClientState(exported.ClientTypeLocalHost, localHostClient))

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
	suite.keeper.SetClientConsensusState(suite.ctx, testClientID, testClientHeight, suite.consensusState)

	nextState := ibctmtypes.ConsensusState{
		Height:             testClientHeight + 5,
		Timestamp:          suite.now,
		Root:               commitmenttypes.NewMerkleRoot([]byte("next")),
		NextValidatorsHash: suite.valSetHash,
	}

	header := ibctmtypes.CreateTestHeader(testClientID, testClientHeight+5, testClientHeight, suite.header.Time.Add(time.Minute),
		suite.valSet, suite.valSet, []tmtypes.PrivValidator{suite.privVal})

	// mock update functionality
	clientState.LatestHeight = uint64(header.Height)
	suite.keeper.SetClientConsensusState(suite.ctx, testClientID, testClientHeight+5, nextState)
	suite.keeper.SetClientState(suite.ctx, testClientID, clientState)

	latest, ok := suite.keeper.GetLatestClientConsensusState(suite.ctx, testClientID)
	suite.Require().True(ok)
	suite.Require().Equal(nextState, latest, "Latest client not returned correctly")

	// Should return existing consensusState at latestClientHeight
	lte, ok := suite.keeper.GetClientConsensusStateLTE(suite.ctx, testClientID, testClientHeight+3)
	suite.Require().True(ok)
	suite.Require().Equal(suite.consensusState, lte, "LTE helper function did not return latest client state below height: %d", testClientHeight+3)
}

func (suite KeeperTestSuite) TestGetAllConsensusStates() {
	expConsensus := []types.ClientConsensusStates{
		types.NewClientConsensusStates(
			testClientID,
			[]exported.ConsensusState{
				ibctmtypes.NewConsensusState(
					suite.consensusState.Timestamp, commitmenttypes.NewMerkleRoot([]byte("hash")), suite.consensusState.GetHeight(), nil,
				),
				ibctmtypes.NewConsensusState(
					suite.consensusState.Timestamp.Add(time.Minute), commitmenttypes.NewMerkleRoot([]byte("app_hash")), suite.consensusState.GetHeight()+1, nil,
				),
			},
		),
		types.NewClientConsensusStates(
			testClientID2,
			[]exported.ConsensusState{
				ibctmtypes.NewConsensusState(
					suite.consensusState.Timestamp.Add(2*time.Minute), commitmenttypes.NewMerkleRoot([]byte("app_hash_2")), suite.consensusState.GetHeight()+2, nil,
				),
			},
		),
	}

	for i := range expConsensus {
		for _, cons := range expConsensus[i].ConsensusStates {
			suite.keeper.SetClientConsensusState(suite.ctx, expConsensus[i].ClientID, cons.GetHeight(), cons)
		}
	}

	consStates := suite.keeper.GetAllConsensusStates(suite.ctx)
	suite.Require().Len(consStates, len(expConsensus))
	suite.Require().Equal(expConsensus, consStates)
}
