package keeper_test

import (
	"testing"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/keeper"
	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	testClientID  = "gaia"
	testClientID2 = "ethbridge"
	testClientID3 = "ethermint"
)

type KeeperTestSuite struct {
	suite.Suite

	cdc            *codec.Codec
	ctx            sdk.Context
	keeper         *keeper.Keeper
	consensusState tendermint.ConsensusState
	header         tendermint.Header
	valSet         *tmtypes.ValidatorSet
	privVal        tmtypes.PrivValidator
}

func (suite *KeeperTestSuite) SetupTest() {
	isCheckTx := false
	app := simapp.Setup(isCheckTx)

	suite.cdc = app.Codec()
	suite.ctx = app.BaseApp.NewContext(isCheckTx, abci.Header{})
	suite.keeper = &app.IBCKeeper.ClientKeeper

	suite.privVal = tmtypes.NewMockPV()

	validator := tmtypes.NewValidator(suite.privVal.GetPubKey(), 1)
	suite.valSet = tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})

	suite.header = tendermint.MakeHeader("gaia", 4, suite.valSet, suite.valSet, []tmtypes.PrivValidator{suite.privVal})

	suite.consensusState = tendermint.ConsensusState{
		Root:             commitment.NewRoot([]byte("hash")),
		ValidatorSetHash: suite.valSet.Hash(),
	}
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestSetClientState() {
	clientState := tendermint.NewClientState(testClientID)
	suite.keeper.SetClientState(suite.ctx, clientState)

	retrievedState, ok := suite.keeper.GetClientState(suite.ctx, testClientID)
	require.True(suite.T(), ok, "GetClientState failed")
	require.Equal(suite.T(), clientState, retrievedState, "Client states are not equal")
}

func (suite *KeeperTestSuite) TestSetClientType() {
	suite.keeper.SetClientType(suite.ctx, testClientID, exported.Tendermint)
	clientType, ok := suite.keeper.GetClientType(suite.ctx, testClientID)

	require.True(suite.T(), ok, "GetClientType failed")
	require.Equal(suite.T(), exported.Tendermint, clientType, "ClientTypes not stored correctly")
}

func (suite *KeeperTestSuite) TestSetConsensusState() {
	suite.keeper.SetConsensusState(suite.ctx, testClientID, suite.consensusState)

	retrievedConsState, ok := suite.keeper.GetConsensusState(suite.ctx, testClientID)

	require.True(suite.T(), ok, "GetConsensusState failed")
	tmConsState, _ := retrievedConsState.(tendermint.ConsensusState)
	// force recalculation of unexported totalVotingPower so we can compare consensusState
	tmConsState.ValidatorSet.TotalVotingPower()
	tmConsState.NextValidatorSet.TotalVotingPower()
	require.Equal(suite.T(), suite.consensusState, tmConsState, "ConsensusState not stored correctly")
}

func (suite KeeperTestSuite) TestGetAllClients() {
	expClients := []exported.ClientState{
		tendermint.NewClientState(testClientID2),
		tendermint.NewClientState(testClientID3),
		tendermint.NewClientState(testClientID),
	}

	for i := range expClients {
		suite.keeper.SetClientState(suite.ctx, expClients[i])
	}

	clients := suite.keeper.GetAllClients(suite.ctx)
	suite.Require().Len(clients, len(expClients))
	suite.Require().Equal(expClients, clients)
}
