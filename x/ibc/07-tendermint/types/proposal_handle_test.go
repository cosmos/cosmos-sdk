package types_test

import (
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
)

var (
	frozenHeight = clienttypes.NewHeight(0, 1)
)

// sanity checks
func (suite *TendermintTestSuite) TestCheckProposedHeaderAndUpdateStateBasic() {
	clientA, _ := suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)
	clientState := suite.chainA.GetClientState(clientA).(*types.ClientState)
	clientStore := suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(suite.chainA.GetContext(), clientA)

	// use nil header
	cs, consState, err := clientState.CheckProposedHeaderAndUpdateState(suite.chainA.GetContext(), suite.chainA.App.AppCodec(), clientStore, nil)
	suite.Require().Error(err)
	suite.Require().Nil(cs)
	suite.Require().Nil(consState)

	clientState.LatestHeight = clientState.LatestHeight.Increment()

	// consensus state for latest height does not exist
	cs, consState, err = clientState.CheckProposedHeaderAndUpdateState(suite.chainA.GetContext(), suite.chainA.App.AppCodec(), clientStore, suite.chainA.LastHeader)
	suite.Require().Error(err)
	suite.Require().Nil(cs)
	suite.Require().Nil(consState)

}

// to expire clients, time needs to be fast forwarded on both chainA and chainB.
// this is to prevent headers from failing when attempting to update later.
func (suite *TendermintTestSuite) TestCheckProposedHeaderAndUpdateState() {
	var (
		clientState *types.ClientState
	)

	testCases := []struct {
		name            string
		malleate        func()
		expPassUnfreeze bool // expected result using a header for unfreezing
		expPassUnexpire bool // expected result using a header for unexpiring
	}{
		{
			"not allowed to be updated, not frozen or expired", func() {
				clientState.AllowUpdateAfterExpiry = false
				clientState.AllowUpdateAfterMisbehaviour = false
			}, false, false,
		},
		{
			"not allowed to be updated, client is frozen", func() {
				clientState.AllowUpdateAfterExpiry = false
				clientState.AllowUpdateAfterMisbehaviour = false

				clientState.FrozenHeight = frozenHeight
			}, false, false,
		},
		{
			"not allowed to be updated, client is expired", func() {
				clientState.AllowUpdateAfterExpiry = false
				clientState.AllowUpdateAfterMisbehaviour = false

				suite.chainA.ExpireClient(clientState.TrustingPeriod)
				suite.chainB.ExpireClient(clientState.TrustingPeriod)
				suite.coordinator.CommitBlock(suite.chainA, suite.chainB)
			}, false, false,
		},

		{
			"not allowed to be updated, client is frozen and expired", func() {
				clientState.AllowUpdateAfterExpiry = false
				clientState.AllowUpdateAfterMisbehaviour = false

				clientState.FrozenHeight = frozenHeight
				suite.chainA.ExpireClient(clientState.TrustingPeriod)
				suite.chainB.ExpireClient(clientState.TrustingPeriod)
				suite.coordinator.CommitBlock(suite.chainA, suite.chainB)
			}, false, false,
		},

		{
			"allowed to be updated only after misbehaviour, not frozen or expired", func() {
				clientState.AllowUpdateAfterExpiry = false
				clientState.AllowUpdateAfterMisbehaviour = true
			}, false, false,
		},

		{
			"PASS: allowed to be updated only after misbehaviour, client is frozen", func() {
				clientState.AllowUpdateAfterExpiry = false
				clientState.AllowUpdateAfterMisbehaviour = true

				clientState.FrozenHeight = frozenHeight
			}, true, false,
		},
		{
			"allowed to be updated only after misbehaviour, client is expired", func() {
				clientState.AllowUpdateAfterExpiry = false
				clientState.AllowUpdateAfterMisbehaviour = true

				suite.chainA.ExpireClient(clientState.TrustingPeriod)
				suite.chainB.ExpireClient(clientState.TrustingPeriod)
				suite.coordinator.CommitBlock(suite.chainA, suite.chainB)
			}, false, false,
		},

		{
			"allowed to be updated only after misbehaviour, client is frozen and expired", func() {
				clientState.AllowUpdateAfterExpiry = false
				clientState.AllowUpdateAfterMisbehaviour = true

				clientState.FrozenHeight = frozenHeight
				suite.chainA.ExpireClient(clientState.TrustingPeriod)
				suite.chainB.ExpireClient(clientState.TrustingPeriod)
				suite.coordinator.CommitBlock(suite.chainA, suite.chainB)
			}, true, true, // client is frozen and expired, lighter validation checks are done
		},
		{
			"allowed to be updated only after expiry, not frozen or expired", func() {
				clientState.AllowUpdateAfterExpiry = true
				clientState.AllowUpdateAfterMisbehaviour = false
			}, false, false,
		},

		{
			"allowed to be updated only after expiry, client is frozen", func() {
				clientState.AllowUpdateAfterExpiry = true
				clientState.AllowUpdateAfterMisbehaviour = false

				clientState.FrozenHeight = frozenHeight
			}, false, false,
		},
		{
			"PASS: allowed to be updated only after expiry, client is expired", func() {
				clientState.AllowUpdateAfterExpiry = true
				clientState.AllowUpdateAfterMisbehaviour = false

				suite.chainA.ExpireClient(clientState.TrustingPeriod)
				suite.chainB.ExpireClient(clientState.TrustingPeriod)
				suite.coordinator.CommitBlock(suite.chainA, suite.chainB)
			}, true, true, // unfreezing headers work since they pass stricter checks
		},

		{
			"allowed to be updated only after expiry, client is frozen and expired", func() {
				clientState.AllowUpdateAfterExpiry = true
				clientState.AllowUpdateAfterMisbehaviour = false

				clientState.FrozenHeight = frozenHeight
				suite.chainA.ExpireClient(clientState.TrustingPeriod)
				suite.chainB.ExpireClient(clientState.TrustingPeriod)
				suite.coordinator.CommitBlock(suite.chainA, suite.chainB)
			}, false, false,
		},
		{
			"allowed to be updated after expiry and misbehaviour, not frozen or expired", func() {
				clientState.AllowUpdateAfterExpiry = true
				clientState.AllowUpdateAfterMisbehaviour = true
			}, false, false,
		},
		{
			"PASS: allowed to be updated after expiry and misbehaviour, client is frozen", func() {
				clientState.AllowUpdateAfterExpiry = true
				clientState.AllowUpdateAfterMisbehaviour = true

				clientState.FrozenHeight = frozenHeight
			}, true, false,
		},
		{
			"PASS: allowed to be updated after expiry and misbehaviour, client is expired", func() {
				clientState.AllowUpdateAfterExpiry = true
				clientState.AllowUpdateAfterMisbehaviour = true

				suite.chainA.ExpireClient(clientState.TrustingPeriod)
				suite.chainB.ExpireClient(clientState.TrustingPeriod)
				suite.coordinator.CommitBlock(suite.chainA, suite.chainB)
			}, true, true, // unfreezing headers work since they pass stricter checks
		},

		{
			"PASS: allowed to be updated after expiry and misbehaviour, client is frozen and expired", func() {
				clientState.AllowUpdateAfterExpiry = true
				clientState.AllowUpdateAfterMisbehaviour = true

				clientState.FrozenHeight = frozenHeight
				suite.chainA.ExpireClient(clientState.TrustingPeriod)
				suite.chainB.ExpireClient(clientState.TrustingPeriod)
				suite.coordinator.CommitBlock(suite.chainA, suite.chainB)
			}, true, true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		// for each test case a header used for unexpiring clients and unfreezing
		// a client are each tested to ensure the work as expected and that
		// unexpiring headers cannot unfreeze clients.
		suite.Run(tc.name, func() {

			// start by testing unexpiring the client
			suite.SetupTest() // reset

			clientA, _ := suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)
			clientState = suite.chainA.GetClientState(clientA).(*types.ClientState)

			tc.malleate()

			// use next header for chainB to unexpire clients but with empty trusted heights
			// and validators. Update chainB time so header won't be expired.
			unexpireClientHeader, err := suite.chainA.ConstructUpdateTMClientHeader(suite.chainB, clientA)
			suite.Require().NoError(err)
			unexpireClientHeader.TrustedHeight = clienttypes.Height{}
			unexpireClientHeader.TrustedValidators = nil

			clientStore := suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(suite.chainA.GetContext(), clientA)
			cs, consState, err := clientState.CheckProposedHeaderAndUpdateState(suite.chainA.GetContext(), suite.chainA.App.AppCodec(), clientStore, unexpireClientHeader)

			if tc.expPassUnexpire {
				suite.Require().NoError(err)
				suite.Require().NotNil(cs)
				suite.Require().NotNil(consState)
			} else {
				suite.Require().Error(err)
			}

			// reset and test unfreezing the client
			suite.SetupTest()

			clientA, _ = suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)
			clientState = suite.chainA.GetClientState(clientA).(*types.ClientState)

			tc.malleate()

			// use next header for chainB to unfreeze client on chainA
			unfreezeClientHeader, err := suite.chainA.ConstructUpdateTMClientHeader(suite.chainB, clientA)
			suite.Require().NoError(err)

			clientStore = suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(suite.chainA.GetContext(), clientA)
			cs, consState, err = clientState.CheckProposedHeaderAndUpdateState(suite.chainA.GetContext(), suite.chainA.App.AppCodec(), clientStore, unfreezeClientHeader)

			if tc.expPassUnfreeze {
				suite.Require().NoError(err)
				suite.Require().Equal(uint64(0), cs.GetFrozenHeight())
				suite.Require().NotNil(consState)
			} else {
				suite.Require().Error(err)
			}

		})
	}
}
