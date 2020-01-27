package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/keeper"
	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

const (
	testClientID  = "gaia"
	testClientID2 = "ethbridge"
	testClientID3 = "ethermint"

	testClientHeight = 5
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
	suite.ctx = app.BaseApp.NewContext(isCheckTx, abci.Header{Height: testClientHeight, ChainID: testClientID})
	suite.keeper = &app.IBCKeeper.ClientKeeper
	suite.privVal = tmtypes.NewMockPV()
	validator := tmtypes.NewValidator(suite.privVal.GetPubKey(), 1)
	suite.valSet = tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})
	suite.header = tendermint.CreateTestHeader(testClientID, testClientHeight, suite.valSet, suite.valSet, []tmtypes.PrivValidator{suite.privVal})
	suite.consensusState = tendermint.ConsensusState{
		Root:             commitment.NewRoot([]byte("hash")),
		ValidatorSetHash: suite.valSet.Hash(),
	}

	var validators staking.Validators
	for i := 1; i < 11; i++ {
		privVal := tmtypes.NewMockPV()
		pk := privVal.GetPubKey()
		val := staking.NewValidator(sdk.ValAddress(pk.Address()), pk, staking.Description{})
		val.Status = sdk.Bonded
		val.Tokens = sdk.NewInt(rand.Int63())
		validators = append(validators, val)

		app.StakingKeeper.SetHistoricalInfo(suite.ctx, int64(i), staking.NewHistoricalInfo(suite.ctx.BlockHeader(), validators))
	}
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestSetClientState() {
	clientState := tendermint.NewClientState(testClientID, testClientHeight)
	suite.keeper.SetClientState(suite.ctx, clientState)

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
	tmConsState, _ := retrievedConsState.(tendermint.ConsensusState)
	require.Equal(suite.T(), suite.consensusState, tmConsState, "ConsensusState not stored correctly")
}

func (suite KeeperTestSuite) TestGetAllClients() {
	expClients := []exported.ClientState{
		tendermint.NewClientState(testClientID2, testClientHeight),
		tendermint.NewClientState(testClientID3, testClientHeight),
		tendermint.NewClientState(testClientID, testClientHeight),
	}

	for i := range expClients {
		suite.keeper.SetClientState(suite.ctx, expClients[i])
	}

	clients := suite.keeper.GetAllClients(suite.ctx)
	suite.Require().Len(clients, len(expClients))
	suite.Require().Equal(expClients, clients)
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
