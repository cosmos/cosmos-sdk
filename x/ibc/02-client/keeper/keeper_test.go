package keeper_test

import (
	"math/rand"
	"testing"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/keeper"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/tendermint"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/cosmos/cosmos-sdk/x/staking"

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
	suite.ctx = app.BaseApp.NewContext(isCheckTx, abci.Header{}).WithBlockHeight(10)
	suite.keeper = &app.IBCKeeper.ClientKeeper

	suite.privVal = tmtypes.NewMockPV()

	validator := tmtypes.NewValidator(suite.privVal.GetPubKey(), 1)
	suite.valSet = tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})

	suite.header = tendermint.MakeHeader("gaia", 4, suite.valSet, suite.valSet, []tmtypes.PrivValidator{suite.privVal})

	suite.consensusState = tendermint.ConsensusState{
		ChainID:          testClientID,
		Height:           3,
		Root:             commitment.NewRoot([]byte("hash")),
		ValidatorSet:     suite.valSet,
		NextValidatorSet: suite.valSet,
	}

	var validators staking.Validators
	var valSets []*tmtypes.ValidatorSet
	for i := 1; i < 11; i++ {
		privVal := tmtypes.NewMockPV()
		pk := privVal.GetPubKey()
		val := staking.NewValidator(sdk.ValAddress(pk.Address()), pk, staking.Description{})
		val.Status = sdk.Bonded
		val.Tokens = sdk.NewInt(rand.Int63())
		validators = append(validators, val)

		valset := tmtypes.NewValidatorSet(validators.ToTmValidators())
		valSets = append(valSets, valset)

		app.StakingKeeper.SetHistoricalInfo(suite.ctx, int64(i), staking.NewHistoricalInfo(suite.ctx.BlockHeader(), validators))
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

func (suite *KeeperTestSuite) TestSetClientConsensusState() {
	suite.keeper.SetClientConsensusState(suite.ctx, testClientID, suite.consensusState)

	retrievedConsState, ok := suite.keeper.GetClientConsensusState(suite.ctx, testClientID)

	require.True(suite.T(), ok, "GetClientConsensusState failed")
	tmConsState, _ := retrievedConsState.(tendermint.ConsensusState)
	// force recalculation of unexported totalVotingPower so we can compare consensusState
	tmConsState.ValidatorSet.TotalVotingPower()
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

func (suite KeeperTestSuite) TestSetCommitter() {
	committer := tendermint.Committer{
		ValidatorSet:   suite.valSet,
		Height:         3,
		NextValSetHash: suite.valSet.Hash(),
	}
	nextCommitter := tendermint.Committer{
		ValidatorSet:   suite.valSet,
		Height:         6,
		NextValSetHash: tmhash.Sum([]byte("next_hash")),
	}

	suite.keeper.SetCommitter(suite.ctx, "gaia", 3, committer)
	suite.keeper.SetCommitter(suite.ctx, "gaia", 6, nextCommitter)

	// fetch the commiter on each respective height
	for i := 0; i < 3; i++ {
		committer, ok := suite.keeper.GetCommitter(suite.ctx, "gaia", uint64(i))
		require.False(suite.T(), ok, "GetCommitter passed on nonexistent height: %d", i)
		require.Nil(suite.T(), committer, "GetCommitter returned committer on nonexistent height: %d", i)
	}

	for i := 3; i < 6; i++ {
		recv, ok := suite.keeper.GetCommitter(suite.ctx, "gaia", uint64(i))
		tmRecv, _ := recv.(tendermint.Committer)
		if tmRecv.ValidatorSet != nil {
			// update validator set's power
			tmRecv.ValidatorSet.TotalVotingPower()
		}
		require.True(suite.T(), ok, "GetCommitter failed on existing height: %d", i)
		require.Equal(suite.T(), committer, recv, "GetCommitter returned committer on nonexistent height: %d", i)
	}

	for i := 6; i < 9; i++ {
		recv, ok := suite.keeper.GetCommitter(suite.ctx, "gaia", uint64(i))
		tmRecv, _ := recv.(tendermint.Committer)
		if tmRecv.ValidatorSet != nil {
			// update validator set's power
			tmRecv.ValidatorSet.TotalVotingPower()
		}
		require.True(suite.T(), ok, "GetCommitter failed on existing height: %d", i)
		require.Equal(suite.T(), nextCommitter, recv, "GetCommitter returned committer on nonexistent height: %d", i)
	}
}

func (suite KeeperTestSuite) TestGetAllClients() {
	expClients := []types.State{
		types.NewClientState(testClientID2),
		types.NewClientState(testClientID3),
		types.NewClientState(testClientID),
	}

	for i := range expClients {
		suite.keeper.SetClientState(suite.ctx, expClients[i])
	}

	clients := suite.keeper.GetAllClients(suite.ctx)
	suite.Require().Len(clients, len(expClients))
	suite.Require().Equal(expClients, clients)
}

func (suite KeeperTestSuite) TestGetConsensusState() {
	cases := []struct {
		name    string
		height  int64
		expPass bool
	}{
		{"latest height", suite.ctx.BlockHeight(), false},
		{"zero height", 0, false},
		{"negative height", -1, false},
		{"height > latest height", suite.ctx.BlockHeight() + 1, false},
		{"latest height - 1", suite.ctx.BlockHeight() - 1, true},
	}

	for i, tc := range cases {
		tc := tc
		cs, found := suite.keeper.GetConsensusState(suite.ctx, tc.height)
		if tc.expPass {
			suite.Require().True(found, "Case %d should have passed: %s", i, tc.name)
			suite.Require().NotNil(cs, "Case %d should have passed: %s", i, tc.name)
		} else {
			suite.Require().False(found, "Case %d should have failed: %s", i, tc.name)
			suite.Require().Nil(cs, "Case %d should have failed: %s", i, tc.name)
		}
	}
}
