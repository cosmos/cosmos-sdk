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
		content *types.ClientUpdateProposal
		err     error
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"valid update client proposal", func() {
				clientA, _ := suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)
				clientState := suite.chainA.GetClientState(clientA)

				tmClientState, ok := clientState.(*ibctmtypes.ClientState)
				suite.Require().True(ok)
				tmClientState.AllowUpdateAfterMisbehaviour = true
				tmClientState.FrozenHeight = tmClientState.LatestHeight
				suite.chainA.App.IBCKeeper.ClientKeeper.SetClientState(suite.chainA.GetContext(), clientA, tmClientState)

				// use next header for chainB to update the client on chainA
				header, err := suite.chainA.ConstructUpdateTMClientHeader(suite.chainB, clientA)
				suite.Require().NoError(err)

				content, err = clienttypes.NewClientUpdateProposal(ibctesting.Title, ibctesting.Description, clientA, header)
				suite.Require().NoError(err)
			}, true,
		},
		{
			"client type does not exist", func() {
				content, err = clienttypes.NewClientUpdateProposal(ibctesting.Title, ibctesting.Description, ibctesting.InvalidID, &ibctmtypes.Header{})
				suite.Require().NoError(err)
			}, false,
		},
		{
			"cannot update localhost", func() {
				content, err = clienttypes.NewClientUpdateProposal(ibctesting.Title, ibctesting.Description, exported.Localhost, &ibctmtypes.Header{})
				suite.Require().NoError(err)
			}, false,
		},
		{
			"client does not exist", func() {
				content, err = clienttypes.NewClientUpdateProposal(ibctesting.Title, ibctesting.Description, ibctesting.InvalidID, &ibctmtypes.Header{})
				suite.Require().NoError(err)
			}, false,
		},
		{
			"cannot unpack header, header is nil", func() {
				clientA, _ := suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)
				content = &clienttypes.ClientUpdateProposal{ibctesting.Title, ibctesting.Description, clientA, nil}
			}, false,
		},
		{
			"update fails", func() {
				header := &ibctmtypes.Header{}
				clientA, _ := suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)
				content, err = clienttypes.NewClientUpdateProposal(ibctesting.Title, ibctesting.Description, clientA, header)
				suite.Require().NoError(err)
			}, false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest() // reset

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
