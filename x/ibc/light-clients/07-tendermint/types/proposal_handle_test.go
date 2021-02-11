package types_test

import (
	"time"

	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

var (
	frozenHeight = clienttypes.NewHeight(0, 1)
)

func (suite *TendermintTestSuite) TestCheckSubstituteUpdateStateBasic() {
	var (
		substitute            string
		substituteClientState exported.ClientState
		initialHeight         clienttypes.Height
	)
	testCases := []struct {
		name     string
		malleate func()
	}{
		{
			"solo machine used for substitute", func() {
				substituteClientState = ibctesting.NewSolomachine(suite.T(), suite.cdc, "solo machine", "", 1).ClientState()
			},
		},
		{
			"initial height and substitute revision numbers do not match", func() {
				initialHeight = clienttypes.NewHeight(substituteClientState.GetLatestHeight().GetRevisionNumber()+1, 1)
			},
		},
		{
			"non-matching substitute", func() {
				substitute, _ := suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)
				substituteClientState = suite.chainA.GetClientState(substitute).(*types.ClientState)
				tmClientState, ok := substituteClientState.(*types.ClientState)
				suite.Require().True(ok)

				tmClientState.ChainId = tmClientState.ChainId + "different chain"
			},
		},
		{
			"updated client is invalid - revision height is zero", func() {
				substitute, _ := suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)
				substituteClientState = suite.chainA.GetClientState(substitute).(*types.ClientState)
				tmClientState, ok := substituteClientState.(*types.ClientState)
				suite.Require().True(ok)
				// match subject
				tmClientState.AllowUpdateAfterMisbehaviour = true
				tmClientState.AllowUpdateAfterExpiry = true

				// will occur. This case should never occur (caught by upstream checks)
				initialHeight = clienttypes.NewHeight(5, 0)
				tmClientState.LatestHeight = clienttypes.NewHeight(5, 0)
			},
		},
		{
			"updated client is expired", func() {
				substitute, _ = suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)
				substituteClientState = suite.chainA.GetClientState(substitute).(*types.ClientState)
				tmClientState, ok := substituteClientState.(*types.ClientState)
				suite.Require().True(ok)
				initialHeight = tmClientState.LatestHeight

				// match subject
				tmClientState.AllowUpdateAfterMisbehaviour = true
				tmClientState.AllowUpdateAfterExpiry = true
				suite.chainA.App.IBCKeeper.ClientKeeper.SetClientState(suite.chainA.GetContext(), substitute, tmClientState)

				// update substitute a few times
				err := suite.coordinator.UpdateClient(suite.chainA, suite.chainB, substitute, exported.Tendermint)
				suite.Require().NoError(err)
				substituteClientState = suite.chainA.GetClientState(substitute)

				err = suite.coordinator.UpdateClient(suite.chainA, suite.chainB, substitute, exported.Tendermint)
				suite.Require().NoError(err)

				suite.chainA.ExpireClient(tmClientState.TrustingPeriod)
				suite.chainB.ExpireClient(tmClientState.TrustingPeriod)
				suite.coordinator.CommitBlock(suite.chainA, suite.chainB)

				substituteClientState = suite.chainA.GetClientState(substitute)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {

			suite.SetupTest() // reset

			subject, _ := suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)
			subjectClientState := suite.chainA.GetClientState(subject).(*types.ClientState)
			subjectClientState.AllowUpdateAfterMisbehaviour = true
			subjectClientState.AllowUpdateAfterExpiry = true

			// expire subject
			suite.chainA.ExpireClient(subjectClientState.TrustingPeriod)
			suite.chainB.ExpireClient(subjectClientState.TrustingPeriod)
			suite.coordinator.CommitBlock(suite.chainA, suite.chainB)

			tc.malleate()

			subjectClientStore := suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(suite.chainA.GetContext(), subject)
			substituteClientStore := suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(suite.chainA.GetContext(), substitute)

			updatedClient, err := subjectClientState.CheckSubstituteAndUpdateState(suite.chainA.GetContext(), suite.chainA.App.AppCodec(), subjectClientStore, substituteClientStore, substituteClientState, initialHeight)
			suite.Require().Error(err)
			suite.Require().Nil(updatedClient)
		})
	}
}

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

			// construct the substitute to match the subject client
			// NOTE: the substitute is explicitly created after the freezing or expiry occurs,
			// primarily to prevent the substitute from becoming frozen. It also should be
			// the natural flow of events in practice. The subject will become frozen/expired
			// and a substitute will be created along with a governance proposal as a response

			substitute, _ := suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)
			substituteClientState := suite.chainA.GetClientState(substitute).(*types.ClientState)
			substituteClientState.AllowUpdateAfterExpiry = tc.AllowUpdateAfterExpiry
			substituteClientState.AllowUpdateAfterMisbehaviour = tc.AllowUpdateAfterMisbehaviour
			suite.chainA.App.IBCKeeper.ClientKeeper.SetClientState(suite.chainA.GetContext(), substitute, substituteClientState)

			initialHeight := substituteClientState.GetLatestHeight()

			// update substitute a few times
			for i := 0; i < 3; i++ {
				err := suite.coordinator.UpdateClient(suite.chainA, suite.chainB, substitute, exported.Tendermint)
				suite.Require().NoError(err)
				// skip a block
				suite.coordinator.CommitBlock(suite.chainA, suite.chainB)
			}

			// get updated substitue
			substituteClientState = suite.chainA.GetClientState(substitute).(*types.ClientState)

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

func (suite *TendermintTestSuite) TestIsMatchingClientState() {
	var (
		subject, substitute                       string
		subjectClientState, substituteClientState *types.ClientState
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"matching clients", func() {
				subjectClientState = suite.chainA.GetClientState(subject).(*types.ClientState)
				substituteClientState = suite.chainA.GetClientState(substitute).(*types.ClientState)
			}, true,
		},
		{
			"matching, frozen height is not used in check for equality", func() {
				subjectClientState.FrozenHeight = frozenHeight
				substituteClientState.FrozenHeight = clienttypes.ZeroHeight()
			}, true,
		},
		{
			"matching, latest height is not used in check for equality", func() {
				subjectClientState.LatestHeight = clienttypes.NewHeight(0, 10)
				substituteClientState.FrozenHeight = clienttypes.ZeroHeight()
			}, true,
		},
		{
			"matching, chain id is different", func() {
				subjectClientState.ChainId = "bitcoin"
				substituteClientState.ChainId = "ethereum"
			}, true,
		},
		{
			"not matching, trusting period is different", func() {
				subjectClientState.TrustingPeriod = time.Duration(time.Hour * 10)
				substituteClientState.TrustingPeriod = time.Duration(time.Hour * 1)
			}, false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest() // reset

			subject, _ = suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)
			substitute, _ = suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)

			tc.malleate()

			suite.Require().Equal(tc.expPass, types.IsMatchingClientState(*subjectClientState, *substituteClientState))

		})
	}
}
