package types_test

import (
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
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

	clientState.LatestHeight = clientState.LatestHeight.Increment().(clienttypes.Height)

	// consensus state for latest height does not exist
	cs, consState, err = clientState.CheckProposedHeaderAndUpdateState(suite.chainA.GetContext(), suite.chainA.App.AppCodec(), clientStore, suite.chainA.LastHeader)
	suite.Require().Error(err)
	suite.Require().Nil(cs)
	suite.Require().Nil(consState)
}

// to expire clients, time needs to be fast forwarded on both chainA and chainB.
// this is to prevent headers from failing when attempting to update later.
func (suite *TendermintTestSuite) TestCheckProposedHeaderAndUpdateState() {
	testCases := []struct {
		name                         string
		AllowUpdateAfterExpiry       bool
		AllowUpdateAfterMisbehaviour bool
		FreezeClient                 bool
		ExpireClient                 bool
		expPassUnfreeze              bool // expected result using a header that passes stronger validation
		expPassUnexpire              bool // expected result using a header that passes weaker validation
	}{
		{
			name:                         "not allowed to be updated, not frozen or expired",
			AllowUpdateAfterExpiry:       false,
			AllowUpdateAfterMisbehaviour: false,
			FreezeClient:                 false,
			ExpireClient:                 false,
			expPassUnfreeze:              false,
			expPassUnexpire:              false,
		},
		{
			name:                         "not allowed to be updated, client is frozen",
			AllowUpdateAfterExpiry:       false,
			AllowUpdateAfterMisbehaviour: false,
			FreezeClient:                 true,
			ExpireClient:                 false,
			expPassUnfreeze:              false,
			expPassUnexpire:              false,
		},
		{
			name:                         "not allowed to be updated, client is expired",
			AllowUpdateAfterExpiry:       false,
			AllowUpdateAfterMisbehaviour: false,
			FreezeClient:                 false,
			ExpireClient:                 true,
			expPassUnfreeze:              false,
			expPassUnexpire:              false,
		},
		{
			name:                         "not allowed to be updated, client is frozen and expired",
			AllowUpdateAfterExpiry:       false,
			AllowUpdateAfterMisbehaviour: false,
			FreezeClient:                 true,
			ExpireClient:                 true,
			expPassUnfreeze:              false,
			expPassUnexpire:              false,
		},
		{
			name:                         "allowed to be updated only after misbehaviour, not frozen or expired",
			AllowUpdateAfterExpiry:       false,
			AllowUpdateAfterMisbehaviour: true,
			FreezeClient:                 false,
			ExpireClient:                 false,
			expPassUnfreeze:              false,
			expPassUnexpire:              false,
		},
		{
			name:                         "PASS: allowed to be updated only after misbehaviour, client is frozen",
			AllowUpdateAfterExpiry:       false,
			AllowUpdateAfterMisbehaviour: true,
			FreezeClient:                 true,
			ExpireClient:                 false,
			expPassUnfreeze:              true,
			expPassUnexpire:              false,
		},
		{
			name:                         "allowed to be updated only after misbehaviour, client is expired",
			AllowUpdateAfterExpiry:       false,
			AllowUpdateAfterMisbehaviour: true,
			FreezeClient:                 false,
			ExpireClient:                 true,
			expPassUnfreeze:              false,
			expPassUnexpire:              false,
		},
		{
			name:                         "allowed to be updated only after misbehaviour, client is frozen and expired",
			AllowUpdateAfterExpiry:       false,
			AllowUpdateAfterMisbehaviour: true,
			FreezeClient:                 true,
			ExpireClient:                 true,
			expPassUnfreeze:              true,
			expPassUnexpire:              true,
		},
		{
			name:                         "allowed to be updated only after expiry, not frozen or expired",
			AllowUpdateAfterExpiry:       true,
			AllowUpdateAfterMisbehaviour: false,
			FreezeClient:                 false,
			ExpireClient:                 false,
			expPassUnfreeze:              false,
			expPassUnexpire:              false,
		},
		{
			name:                         "allowed to be updated only after expiry, client is frozen",
			AllowUpdateAfterExpiry:       true,
			AllowUpdateAfterMisbehaviour: false,
			FreezeClient:                 true,
			ExpireClient:                 false,
			expPassUnfreeze:              false,
			expPassUnexpire:              false,
		},
		{
			name:                         "PASS: allowed to be updated only after expiry, client is expired",
			AllowUpdateAfterExpiry:       true,
			AllowUpdateAfterMisbehaviour: false,
			FreezeClient:                 false,
			ExpireClient:                 true,
			expPassUnfreeze:              true,
			expPassUnexpire:              true,
		},
		{
			name:                         "allowed to be updated only after expiry, client is frozen and expired",
			AllowUpdateAfterExpiry:       true,
			AllowUpdateAfterMisbehaviour: false,
			FreezeClient:                 true,
			ExpireClient:                 true,
			expPassUnfreeze:              false,
			expPassUnexpire:              false,
		},
		{
			name:                         "allowed to be updated after expiry and misbehaviour, not frozen or expired",
			AllowUpdateAfterExpiry:       true,
			AllowUpdateAfterMisbehaviour: true,
			FreezeClient:                 false,
			ExpireClient:                 false,
			expPassUnfreeze:              false,
			expPassUnexpire:              false,
		},
		{
			name:                         "PASS: allowed to be updated after expiry and misbehaviour, client is frozen",
			AllowUpdateAfterExpiry:       true,
			AllowUpdateAfterMisbehaviour: true,
			FreezeClient:                 true,
			ExpireClient:                 false,
			expPassUnfreeze:              true,
			expPassUnexpire:              false,
		},
		{
			name:                         "PASS: allowed to be updated after expiry and misbehaviour, client is expired",
			AllowUpdateAfterExpiry:       true,
			AllowUpdateAfterMisbehaviour: true,
			FreezeClient:                 false,
			ExpireClient:                 true,
			expPassUnfreeze:              true,
			expPassUnexpire:              true,
		},
		{
			name:                         "PASS: allowed to be updated after expiry and misbehaviour, client is frozen and expired",
			AllowUpdateAfterExpiry:       true,
			AllowUpdateAfterMisbehaviour: true,
			FreezeClient:                 true,
			ExpireClient:                 true,
			expPassUnfreeze:              true,
			expPassUnexpire:              true,
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

			// construct client state based on test case parameters
			clientA, _ := suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)
			clientState := suite.chainA.GetClientState(clientA).(*types.ClientState)
			clientState.AllowUpdateAfterExpiry = tc.AllowUpdateAfterExpiry
			clientState.AllowUpdateAfterMisbehaviour = tc.AllowUpdateAfterMisbehaviour
			if tc.FreezeClient {
				clientState.FrozenHeight = frozenHeight
			}
			if tc.ExpireClient {
				suite.chainA.ExpireClient(clientState.TrustingPeriod)
				suite.chainB.ExpireClient(clientState.TrustingPeriod)
				suite.coordinator.CommitBlock(suite.chainA, suite.chainB)
			}

			// use next header for chainB to unfreeze client on chainA
			unfreezeClientHeader, err := suite.chainA.ConstructUpdateTMClientHeader(suite.chainB, clientA)
			suite.Require().NoError(err)

			clientStore := suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(suite.chainA.GetContext(), clientA)
			cs, consState, err := clientState.CheckProposedHeaderAndUpdateState(suite.chainA.GetContext(), suite.chainA.App.AppCodec(), clientStore, unfreezeClientHeader)

			if tc.expPassUnfreeze {
				suite.Require().NoError(err)
				suite.Require().Equal(clienttypes.ZeroHeight(), cs.GetFrozenHeight())
				suite.Require().NotNil(consState)
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(cs)
				suite.Require().Nil(consState)
			}

			// use next header for chainB to unexpire clients but with empty trusted heights
			// and validators. Update chainB time so header won't be expired.
			unexpireClientHeader, err := suite.chainA.ConstructUpdateTMClientHeader(suite.chainB, clientA)
			suite.Require().NoError(err)
			unexpireClientHeader.TrustedHeight = clienttypes.ZeroHeight()
			unexpireClientHeader.TrustedValidators = nil

			clientStore = suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(suite.chainA.GetContext(), clientA)
			cs, consState, err = clientState.CheckProposedHeaderAndUpdateState(suite.chainA.GetContext(), suite.chainA.App.AppCodec(), clientStore, unexpireClientHeader)

			if tc.expPassUnexpire {
				suite.Require().NoError(err)
				suite.Require().NotNil(cs)
				suite.Require().NotNil(consState)
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(cs)
				suite.Require().Nil(consState)
			}
		})
	}
}

// test softer validation on headers used for unexpiring clients
func (suite *TendermintTestSuite) TestCheckProposedHeader() {
	var (
		header      *types.Header
		clientState *types.ClientState
		clientA     string
		err         error
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
			"invalid signed header", func() {
				header.SignedHeader = nil
			}, false,
		},
		{
			"header time is less than or equal to consensus state timestamp", func() {
				consensusState, found := suite.chainA.GetConsensusState(clientA, clientState.GetLatestHeight())
				suite.Require().True(found)
				consensusState.(*types.ConsensusState).Timestamp = header.GetTime()
				suite.chainA.App.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.chainA.GetContext(), clientA, clientState.GetLatestHeight(), consensusState)

				// update block time so client is expired
				suite.chainA.ExpireClient(clientState.TrustingPeriod)
				suite.chainB.ExpireClient(clientState.TrustingPeriod)
				suite.coordinator.CommitBlock(suite.chainA, suite.chainB)
			}, false,
		},
		{
			"header height is not newer than client state", func() {
				consensusState, found := suite.chainA.GetConsensusState(clientA, clientState.GetLatestHeight())
				suite.Require().True(found)
				clientState.LatestHeight = header.GetHeight().(clienttypes.Height)
				suite.chainA.App.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.chainA.GetContext(), clientA, clientState.GetLatestHeight(), consensusState)

			}, false,
		},
		{
			"signed header failed validate basic - wrong chain ID", func() {
				clientState.ChainId = ibctesting.InvalidID
			}, false,
		},
		{
			"header is already expired", func() {
				// expire client
				suite.chainA.ExpireClient(clientState.TrustingPeriod)
				suite.chainB.ExpireClient(clientState.TrustingPeriod)
				suite.coordinator.CommitBlock(suite.chainA, suite.chainB)
			}, false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest() // reset

			clientA, _ = suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)
			clientState = suite.chainA.GetClientState(clientA).(*types.ClientState)
			clientState.AllowUpdateAfterExpiry = true
			clientState.AllowUpdateAfterMisbehaviour = false

			// expire client
			suite.chainA.ExpireClient(clientState.TrustingPeriod)
			suite.chainB.ExpireClient(clientState.TrustingPeriod)
			suite.coordinator.CommitBlock(suite.chainA, suite.chainB)

			// use next header for chainB to unexpire clients but with empty trusted heights
			// and validators.
			header, err = suite.chainA.ConstructUpdateTMClientHeader(suite.chainB, clientA)
			suite.Require().NoError(err)
			header.TrustedHeight = clienttypes.ZeroHeight()
			header.TrustedValidators = nil

			tc.malleate()

			clientStore := suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(suite.chainA.GetContext(), clientA)
			cs, consState, err := clientState.CheckProposedHeaderAndUpdateState(suite.chainA.GetContext(), suite.chainA.App.AppCodec(), clientStore, header)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(cs)
				suite.Require().NotNil(consState)
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(cs)
				suite.Require().Nil(consState)
			}
		})
	}
}
