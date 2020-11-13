package types_test

import (
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	solomachinetypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/06-solomachine/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
)

func (suite *TendermintTestSuite) TestGetConsensusState() {
	var (
		height  exported.Height
		clientA string
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"success", func() {}, true,
		},
		{
			"consensus state not found", func() {
				// use height with no consensus state set
				height = height.(clienttypes.Height).Increment()
			}, false,
		},
		{
			"not a consensus state interface", func() {
				// marshal an empty client state and set as consensus state
				store := suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(suite.chainA.GetContext(), clientA)
				clientStateBz := suite.chainA.App.IBCKeeper.ClientKeeper.MustMarshalClientState(&types.ClientState{})
				store.Set(host.ConsensusStateKey(height), clientStateBz)
			}, false,
		},
		{
			"invalid consensus state (solomachine)", func() {
				// marshal and set solomachine consensus state
				store := suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(suite.chainA.GetContext(), clientA)
				consensusStateBz := suite.chainA.App.IBCKeeper.ClientKeeper.MustMarshalConsensusState(&solomachinetypes.ConsensusState{})
				store.Set(host.ConsensusStateKey(height), consensusStateBz)
			}, false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()

			clientA, _, _, _, _, _ = suite.coordinator.Setup(suite.chainA, suite.chainB, channeltypes.UNORDERED)
			clientState := suite.chainA.GetClientState(clientA)
			height = clientState.GetLatestHeight()

			tc.malleate() // change vars as necessary

			store := suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(suite.chainA.GetContext(), clientA)
			consensusState, err := types.GetConsensusState(store, suite.chainA.Codec, height)

			if tc.expPass {
				suite.Require().NoError(err)
				expConsensusState, found := suite.chainA.GetConsensusState(clientA, height)
				suite.Require().True(found)
				suite.Require().Equal(expConsensusState, consensusState)
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(consensusState)
			}
		})
	}
}
