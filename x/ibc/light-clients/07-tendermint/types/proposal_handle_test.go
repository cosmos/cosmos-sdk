package types_test

import (
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
	//	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

var (
	frozenHeight = clienttypes.NewHeight(0, 1)
)

/*
// TODO update to handle other lines of code
// sanity checks
func (suite *TendermintTestSuite) TestCheckSubstituteUpdateStateBasic() {
	clientA, _ := suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)
	clientState := suite.chainA.GetClientState(clientA).(*types.ClientState)
	clientStore := suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(suite.chainA.GetContext(), clientA)

	// use nil header
	cs, consState, err := clientState.CheckProposedHeaderAndUpdateState(suite.chainA.GetContext(), suite.chainA.App.AppCodec(), clientStore, nil)
	suite.Require().Error(err)
	suite.Require().Nil(cs)
	suite.Require().Nil(consState)

	clientState.LatestHeight = clientState.LatestHeight.Increment().(clienttypes.Height)

	// consensus state for latest height does not exist
	cs, consState, err = clientState.CheckProposedHeaderAndUpdateState(suite.chainA.GetContext(), suite.chainA.App.AppCodec(), clientStore, suite.chainA.LastHeader)
	suite.Require().Error(err)
	suite.Require().Nil(cs)
	suite.Require().Nil(consState)
}
*/

// to expire clients, time needs to be fast forwarded on both chainA and chainB.
// this is to prevent headers from failing when attempting to update later.
func (suite *TendermintTestSuite) TestCheckSubstituteAndUpdateState() {
	testCases := []struct {
		name                         string
		AllowUpdateAfterExpiry       bool
		AllowUpdateAfterMisbehaviour bool
		FreezeClient                 bool
		ExpireClient                 bool
		expPass                      bool
	}{
		{
			name:                         "not allowed to be updated, not frozen or expired",
			AllowUpdateAfterExpiry:       false,
			AllowUpdateAfterMisbehaviour: false,
			FreezeClient:                 false,
			ExpireClient:                 false,
			expPass:                      false,
		},
		{
			name:                         "not allowed to be updated, client is frozen",
			AllowUpdateAfterExpiry:       false,
			AllowUpdateAfterMisbehaviour: false,
			FreezeClient:                 true,
			ExpireClient:                 false,
			expPass:                      false,
		},
		{
			name:                         "not allowed to be updated, client is expired",
			AllowUpdateAfterExpiry:       false,
			AllowUpdateAfterMisbehaviour: false,
			FreezeClient:                 false,
			ExpireClient:                 true,
			expPass:                      false,
		},
		{
			name:                         "not allowed to be updated, client is frozen and expired",
			AllowUpdateAfterExpiry:       false,
			AllowUpdateAfterMisbehaviour: false,
			FreezeClient:                 true,
			ExpireClient:                 true,
			expPass:                      false,
		},
		{
			name:                         "allowed to be updated only after misbehaviour, not frozen or expired",
			AllowUpdateAfterExpiry:       false,
			AllowUpdateAfterMisbehaviour: true,
			FreezeClient:                 false,
			ExpireClient:                 false,
			expPass:                      false,
		},
		{
			name:                         "PASS: allowed to be updated only after misbehaviour, client is frozen",
			AllowUpdateAfterExpiry:       false,
			AllowUpdateAfterMisbehaviour: true,
			FreezeClient:                 true,
			ExpireClient:                 false,
			expPass:                      true,
		},
		{
			name:                         "allowed to be updated only after misbehaviour, client is expired",
			AllowUpdateAfterExpiry:       false,
			AllowUpdateAfterMisbehaviour: true,
			FreezeClient:                 false,
			ExpireClient:                 true,
			expPass:                      false,
		},
		{
			name:                         "PASS: allowed to be updated only after misbehaviour, client is frozen and expired",
			AllowUpdateAfterExpiry:       false,
			AllowUpdateAfterMisbehaviour: true,
			FreezeClient:                 true,
			ExpireClient:                 true,
			expPass:                      true,
		},
		{
			name:                         "allowed to be updated only after expiry, not frozen or expired",
			AllowUpdateAfterExpiry:       true,
			AllowUpdateAfterMisbehaviour: false,
			FreezeClient:                 false,
			ExpireClient:                 false,
			expPass:                      false,
		},
		{
			name:                         "allowed to be updated only after expiry, client is frozen",
			AllowUpdateAfterExpiry:       true,
			AllowUpdateAfterMisbehaviour: false,
			FreezeClient:                 true,
			ExpireClient:                 false,
			expPass:                      false,
		},
		{
			name:                         "PASS: allowed to be updated only after expiry, client is expired",
			AllowUpdateAfterExpiry:       true,
			AllowUpdateAfterMisbehaviour: false,
			FreezeClient:                 false,
			ExpireClient:                 true,
			expPass:                      true,
		},
		{
			name:                         "allowed to be updated only after expiry, client is frozen and expired",
			AllowUpdateAfterExpiry:       true,
			AllowUpdateAfterMisbehaviour: false,
			FreezeClient:                 true,
			ExpireClient:                 true,
			expPass:                      false,
		},
		{
			name:                         "allowed to be updated after expiry and misbehaviour, not frozen or expired",
			AllowUpdateAfterExpiry:       true,
			AllowUpdateAfterMisbehaviour: true,
			FreezeClient:                 false,
			ExpireClient:                 false,
			expPass:                      false,
		},
		{
			name:                         "PASS: allowed to be updated after expiry and misbehaviour, client is frozen",
			AllowUpdateAfterExpiry:       true,
			AllowUpdateAfterMisbehaviour: true,
			FreezeClient:                 true,
			ExpireClient:                 false,
			expPass:                      true,
		},
		{
			name:                         "PASS: allowed to be updated after expiry and misbehaviour, client is expired",
			AllowUpdateAfterExpiry:       true,
			AllowUpdateAfterMisbehaviour: true,
			FreezeClient:                 false,
			ExpireClient:                 true,
			expPass:                      true,
		},
		{
			name:                         "PASS: allowed to be updated after expiry and misbehaviour, client is frozen and expired",
			AllowUpdateAfterExpiry:       true,
			AllowUpdateAfterMisbehaviour: true,
			FreezeClient:                 true,
			ExpireClient:                 true,
			expPass:                      true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		// for each test case a header used for unexpiring clients and unfreezing
		// a client are each tested to ensure that unexpiry headers cannot update
		// a client when a unfreezing header is required.
		suite.Run(tc.name, func() {

			// start by testing unexpiring the client
			suite.SetupTest() // reset

			// construct subject using test case parameters
			subject, _ := suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)
			subjectClientState := suite.chainA.GetClientState(subject).(*types.ClientState)
			subjectClientState.AllowUpdateAfterExpiry = tc.AllowUpdateAfterExpiry
			subjectClientState.AllowUpdateAfterMisbehaviour = tc.AllowUpdateAfterMisbehaviour

			// apply freezing or expiry as determined by the test case
			if tc.FreezeClient {
				subjectClientState.FrozenHeight = frozenHeight
			}
			if tc.ExpireClient {
				suite.chainA.ExpireClient(subjectClientState.TrustingPeriod)
				suite.chainB.ExpireClient(subjectClientState.TrustingPeriod)
				suite.coordinator.CommitBlock(suite.chainA, suite.chainB)
			}

			// cosntruct the substitute to match the subject client
			// NOTE: the substitute is explicitly created after the freezing or expiry occurs,
			// primarily to prevent the substitute from becoming frozen. It also should be
			// the natural flow of events in practice. The subject will become frozne/expired
			// and a substitute will be created along with a governance proposal as a response

			substitute, _ := suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)
			substituteClientState := suite.chainA.GetClientState(substitute).(*types.ClientState)
			substituteClientState.AllowUpdateAfterExpiry = tc.AllowUpdateAfterExpiry
			substituteClientState.AllowUpdateAfterMisbehaviour = tc.AllowUpdateAfterMisbehaviour
			initialHeight := substituteClientState.GetLatestHeight()

			// update substitute a few times
			suite.coordinator.UpdateClient(suite.chainA, suite.chainB, substitute, exported.Tendermint)
			suite.coordinator.UpdateClient(suite.chainA, suite.chainB, substitute, exported.Tendermint)
			suite.coordinator.UpdateClient(suite.chainA, suite.chainB, substitute, exported.Tendermint)

			subjectClientStore := suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(suite.chainA.GetContext(), subject)
			substituteClientStore := suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(suite.chainA.GetContext(), substitute)
			updatedClient, err := subjectClientState.CheckSubstituteAndUpdateState(suite.chainA.GetContext(), suite.chainA.App.AppCodec(), subjectClientStore, substituteClientStore, substituteClientState, initialHeight)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(clienttypes.ZeroHeight(), updatedClient.GetFrozenHeight())
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(updatedClient)
			}

		})
	}
}

// TODO: add test for equality helper function
