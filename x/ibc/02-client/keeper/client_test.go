package keeper_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/tendermint"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/types"
)

const (
	testClientType exported.ClientType = iota + 2
)

func (suite *KeeperTestSuite) TestCreateClient() {
	// Test Valid CreateClient
	state, err := suite.keeper.CreateClient(suite.ctx, testClientID, exported.Tendermint, suite.consensusState)
	require.Nil(suite.T(), err, "CreateClient failed")

	// Test ClientState stored correctly
	expectedState := types.State{
		ID:     testClientID,
		Frozen: false,
	}
	require.Equal(suite.T(), expectedState, state, "Incorrect ClientState returned")

	// Test ClientType and VerifiedRoot stored correctly
	clientType, _ := suite.keeper.GetClientType(suite.ctx, testClientID)
	require.Equal(suite.T(), exported.Tendermint, clientType, "Incorrect ClientType stored")
	root, _ := suite.keeper.GetVerifiedRoot(suite.ctx, testClientID, suite.consensusState.GetHeight())
	require.Equal(suite.T(), suite.consensusState.GetRoot(), root, "Incorrect root stored")

	// Test that trying to CreateClient on existing client fails
	_, err = suite.keeper.CreateClient(suite.ctx, testClientID, exported.Tendermint, suite.consensusState)
	require.NotNil(suite.T(), err, "CreateClient on existing client: %s passed", testClientID)
}

func (suite *KeeperTestSuite) TestUpdateClient() {
	privVal := tmtypes.NewMockPV()
	validator := tmtypes.NewValidator(privVal.GetPubKey(), 1)
	altValSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})
	altSigners := []tmtypes.PrivValidator{privVal}

	// Test invalid cases all fail and do not update state
	cases := []struct {
		name     string
		malleate func()
		expErr   bool
	}{
		{"valid update", func() {}, false},
		{"wrong client type", func() {
			suite.keeper.SetClientType(suite.ctx, testClientID, testClientType)
		}, true},
		{"frozen client", func() {
			clientState, _ := suite.keeper.GetClientState(suite.ctx, testClientID)
			clientState.Frozen = true
			suite.keeper.SetClientState(suite.ctx, clientState)
		}, true},
		{"past height", func() {
			suite.header = tendermint.MakeHeader(2, suite.valSet, suite.valSet, []tmtypes.PrivValidator{suite.privVal})
		}, true},
		{"validatorHash incorrect", func() {
			suite.header = tendermint.MakeHeader(4, altValSet, suite.valSet, altSigners)
		}, true},
		{"nextHash incorrect", func() {
			suite.header.NextValidatorSet = altValSet
		}, true},
		{"header fails validateBasic", func() {
			suite.header.ChainID = "test"
		}, true},
		{"verify future commit fails", func() {
			suite.consensusState.NextValidatorSet = altValSet
			suite.keeper.SetConsensusState(suite.ctx, testClientID, suite.consensusState)
		}, true},
	}

	for _, tc := range cases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			// Reset suite on each subtest
			suite.SetupTest()

			_, err := suite.keeper.CreateClient(suite.ctx, testClientID, exported.Tendermint, suite.consensusState)
			require.Nil(suite.T(), err, "CreateClient failed")

			tc.malleate()
			err = suite.keeper.UpdateClient(suite.ctx, testClientID, suite.header)

			retrievedConsState, _ := suite.keeper.GetConsensusState(suite.ctx, testClientID)
			tmConsState, _ := retrievedConsState.(tendermint.ConsensusState)
			tmConsState.NextValidatorSet.TotalVotingPower()
			retrievedRoot, _ := suite.keeper.GetVerifiedRoot(suite.ctx, testClientID, suite.consensusState.GetHeight()+1)
			if tc.expErr {
				require.NotNil(suite.T(), err, "Invalid UpdateClient passed", tc.name)

				// require no state changes occurred
				require.Equal(suite.T(), suite.consensusState, tmConsState, "Consensus state changed after invalid UpdateClient")
				require.Nil(suite.T(), retrievedRoot, "Root added for new height after invalid UpdateClient")
			} else {
				require.Nil(suite.T(), err, "Valid UpdateClient failed", tc.name)

				// require state changes were performed correctly
				require.Equal(suite.T(), suite.header.GetHeight(), retrievedConsState.GetHeight(), "height not updated correctly")
				require.Equal(suite.T(), commitment.NewRoot(suite.header.AppHash), retrievedConsState.GetRoot(), "root not updated correctly")
				require.Equal(suite.T(), suite.header.NextValidatorSet, tmConsState.NextValidatorSet, "NextValidatorSet not updated correctly")

			}

		})
	}
}
