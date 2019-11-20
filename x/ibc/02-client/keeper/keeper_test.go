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
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/tendermint"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	testClientID = "gaia"
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

	suite.header = tendermint.MakeHeader(4, suite.valSet, suite.valSet, []tmtypes.PrivValidator{suite.privVal})

	suite.consensusState = tendermint.ConsensusState{
		ChainID:          testClientID,
		Height:           3,
		Root:             commitment.NewRoot([]byte("hash")),
		NextValidatorSet: suite.valSet,
	}
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestSetClientState() {
	clientState := types.NewClientState(testClientID)
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
	tmConsState.NextValidatorSet.TotalVotingPower()
	require.Equal(suite.T(), suite.consensusState, tmConsState, "ConsensusState not stored correctly")
}

func (suite *KeeperTestSuite) TestSetVerifiedRoot() {
	root := commitment.NewRoot([]byte("hash"))
	suite.keeper.SetVerifiedRoot(suite.ctx, testClientID, 3, root)

	retrievedRoot, ok := suite.keeper.GetVerifiedRoot(suite.ctx, testClientID, 3)

	require.True(suite.T(), ok, "GetVerifiedRoot failed")
	require.Equal(suite.T(), root, retrievedRoot, "Root stored incorrectly")
}
