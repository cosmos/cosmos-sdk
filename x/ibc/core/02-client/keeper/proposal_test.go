package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

func (suite *KeeperTestSuite) TestClientUpdateProposal() {
	var (
		subject, substitute                       string
		subjectClientState, substituteClientState exported.ClientState
		initialHeight                             clienttypes.Height
		content                                   *types.ClientUpdateProposal
		err                                       error
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"valid update client proposal", func() {
				tmClientState, ok := subjectClientState.(*ibctmtypes.ClientState)
				suite.Require().True(ok)
				tmClientState.AllowUpdateAfterMisbehaviour = true
				tmClientState.FrozenHeight = tmClientState.LatestHeight
				suite.chainA.App.IBCKeeper.ClientKeeper.SetClientState(suite.chainA.GetContext(), subject, tmClientState)

				// replicate changes to substitute (they must match)
				tmClientState, ok = substituteClientState.(*ibctmtypes.ClientState)
				suite.Require().True(ok)
				tmClientState.AllowUpdateAfterMisbehaviour = true
				suite.chainA.App.IBCKeeper.ClientKeeper.SetClientState(suite.chainA.GetContext(), substitute, tmClientState)

				content = clienttypes.NewClientUpdateProposal(ibctesting.Title, ibctesting.Description, subject, substitute, initialHeight)
			}, true,
		},
		{
			"cannot use localhost as subject", func() {
				content = clienttypes.NewClientUpdateProposal(ibctesting.Title, ibctesting.Description, exported.Localhost, substitute, initialHeight)
			}, false,
		},
		{
			"cannot use localhost as substitute", func() {
				content = clienttypes.NewClientUpdateProposal(ibctesting.Title, ibctesting.Description, subject, exported.Localhost, initialHeight)
			}, false,
		},
		{
			"subject client does not exist", func() {
				content = clienttypes.NewClientUpdateProposal(ibctesting.Title, ibctesting.Description, ibctesting.InvalidID, substitute, initialHeight)
			}, false,
		},
		{
			"substitute client does not exist", func() {
				content = clienttypes.NewClientUpdateProposal(ibctesting.Title, ibctesting.Description, subject, ibctesting.InvalidID, initialHeight)
			}, false,
		},
		{
			"subject and substitute have equal latest height", func() {
				tmClientState, ok := subjectClientState.(*ibctmtypes.ClientState)
				suite.Require().True(ok)
				tmClientState.LatestHeight = substituteClientState.GetLatestHeight().(clienttypes.Height)
				suite.chainA.App.IBCKeeper.ClientKeeper.SetClientState(suite.chainA.GetContext(), subject, tmClientState)

				content = clienttypes.NewClientUpdateProposal(ibctesting.Title, ibctesting.Description, subject, substitute, initialHeight)
			}, false,
		},

		{
			"subject and substitute use different revision numbers", func() {
				tmClientState, ok := substituteClientState.(*ibctmtypes.ClientState)
				suite.Require().True(ok)
				tmClientState.LatestHeight = clienttypes.NewHeight(tmClientState.GetLatestHeight().GetRevisionNumber()+1, tmClientState.GetLatestHeight().GetRevisionHeight())
				suite.chainA.App.IBCKeeper.ClientKeeper.SetClientState(suite.chainA.GetContext(), substitute, tmClientState)

				content = clienttypes.NewClientUpdateProposal(ibctesting.Title, ibctesting.Description, subject, substitute, initialHeight)
			}, false,
		},
		{
			"update fails, client is not frozen or expired", func() {
				content = clienttypes.NewClientUpdateProposal(ibctesting.Title, ibctesting.Description, subject, substitute, initialHeight)
			}, false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest() // reset

			subject, _ = suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)
			subjectClientState = suite.chainA.GetClientState(subject)
			substitute, _ = suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)
			initialHeight = clienttypes.NewHeight(subjectClientState.GetLatestHeight().GetRevisionNumber(), subjectClientState.GetLatestHeight().GetRevisionHeight()+1)

			// update substitute twice
			suite.coordinator.UpdateClient(suite.chainA, suite.chainB, substitute, exported.Tendermint)
			suite.coordinator.UpdateClient(suite.chainA, suite.chainB, substitute, exported.Tendermint)
			substituteClientState = suite.chainA.GetClientState(substitute)

			tc.malleate()

			err = suite.chainA.App.IBCKeeper.ClientKeeper.ClientUpdateProposal(suite.chainA.GetContext(), content)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}

}
