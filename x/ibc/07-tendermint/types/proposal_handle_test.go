package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

func (suite *TendermintTestSuite) TestCheckProposedHeaderAndUpdateState() {
	testCases := []struct {
		name            string
		malleate        func()
		expPassUnfreeze bool // expected result using a header for unfreezing
		expPassUnexpire bool // expected result using a header for unexpiring
	}{
		{
			"invalid header - solo machine", func() {

			}, false, false,
		},
		{
			"consnesus state not found", func() {

			}, false, false,
		},

		{
			"not allowed to be updated, not frozen or expired", func() {

			}, false, false,
		},
		{
			"not allowed to be updated, client is frozen", func() {

			}, false, false,
		},
		{
			"not allowed to be updated, client is expired", func() {

			}, false, false,
		},

		{
			"not allowed to be updated, client is frozen and expired", func() {

			}, false, false,
		},

		{
			"allowed to be updated only after misbehaviour, not frozen or expired", func() {

			}, false, false,
		},

		{
			"PASS: allowed to be updated only after misbehaviour, client is frozen", func() {

			}, true, false,
		},
		{
			"allowed to be updated only after misbehaviour, client is expired", func() {

			}, false, false,
		},

		{
			"allowed to be updated only after misbehaviour, client is frozen and expired", func() {

			}, false, false,
		},
		{
			"allowed to be updated only after expiry, not frozen or expired", func() {

			}, false, false,
		},

		{
			"allowed to be updated only after expiry, client is frozen", func() {

			}, false, false,
		},
		{
			"PASS: allowed to be updated only after expiry, client is expired", func() {

			}, true, true, // unfreezing headers work since they pass stricter checks
		},

		{
			"allowed to be updated only after expiry, client is frozen and expired", func() {

			}, false, false,
		},
		{
			"allowed to be updated after expiry and misbehaviour, not frozen or expired", func() {

			}, false, false,
		},
		{
			"PASS: allowed to be updated after expiry and misbehaviour, client is frozen", func() {

			}, true, false,
		},
		{
			"PASS: allowed to be updated after expiry and misbehaviour, client is expired", func() {

			}, true, true, // unfreezing headers work since they pass stricter checks
		},

		{
			"PASS: allowed to be updated after expiry and misbehaviour, client is frozen and expired", func() {

			}, true, true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest() // reset

			unexpiredHeader := types.Header{}
			unfreezeHeader := types.Header{}

			cs, consState, err := cs.CheckProposedHeaderAndUpdateState(suite.chainA.GetContext(), suite.chainA.App.AppCodec(), suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(suite.chainA.GetContext(), clientA), unexpiredHeader)

			if tc.expPassUnexpire {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}

			cs, consState, err = cs.CheckProposedHeaderAndUpdateState(suite.chainA.GetContext(), suite.chainA.App.AppCodec(), suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(suite.chainA.GetContext(), clientA), unfreezeHeader)

			if tc.expPassUnfreeze {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}

		})
	}
}

func (suite *TendermintTestSuite) testClientUpdateProposal(clientID string, header *ibctmtypes.Header) *clienttypes.ClientUpdateProposal {
	proposal, err := clienttypes.NewClientUpdateProposal(ibctesting.Title, ibctesting.Description, clientID, header)
	suite.Require().NoError(err)
	return proposal
}

func (suite *ProposalHandlerTestSuite) testStoredClientStatus(exptedHeight uint64, expectedIsFrozen bool, expectedLatestTimestamp uint64, clientID string) {
	ibcKeeper := *suite.app.IBCKeeper
	clientkeeper := ibcKeeper.ClientKeeper
	clientStatus, ok := clientkeeper.GetClientState(suite.ctx, clientID)
	suite.Require().True(ok)
	suite.Require().Equal(exptedHeight, clientStatus.GetLatestHeight())
	suite.Require().Equal(expectedIsFrozen, clientStatus.IsFrozen())
}

func (suite *ProposalHandlerTestSuite) TestClientUpdateProposalHandler() {
	expiredTime := latestTimestamp.Add(trustingPeriod).Add(time.Minute)
	expiredCtx := suite.app.BaseApp.NewContext(false, tmproto.Header{Time: expiredTime})

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			// Overwrite the suite.ctx with the test ctx.
			// The test ctx is specific for each test cases.
			testCtx := suite.ctx.WithBlockTime(tc.ctx.BlockTime())

			suite.Require().Equal(tc.clientIsExpired, tc.clientState.Expired(testCtx.BlockTime()))
			suite.Require().Equal(tc.clientIsFrozen, tc.clientState.IsFrozen())

			ibcKeeper := *suite.app.IBCKeeper
			clientkeeper := ibcKeeper.ClientKeeper

			// create test client
			_, err := clientkeeper.CreateClient(testCtx, clientID, tc.clientState, suite.consensusState)
			suite.Require().NoError(err)

			// check stored clientStatus before update
			suite.testStoredClientStatus(height, tc.clientState.IsFrozen(), tc.clientState.GetLatestTimestamp(), clientID)

			// create proposal to update the client with the new header (suite.header)
			p, err := testClientUpdateProposal(clientID, suite.header)
			suite.Require().NoError(err)

			// handle client proposal
			hdlr := client.NewClientUpdateProposalHandler(ibcKeeper.ClientKeeper)
			err = hdlr(testCtx, p)

			if tc.expPass {
				suite.Require().NoError(err)

				isFrozen := tc.clientState.IsFrozen()
				if tc.clientState.IsFrozen() && tc.clientState.AllowGovernanceOverrideAfterMisbehaviour {
					isFrozen = false
				}
				// check stored clientStatus after update
				suite.testStoredClientStatus(height+1, isFrozen, uint64(suite.header.GetTime().UnixNano()), clientID)

			} else {
				suite.Require().Error(err, err)
			}
		})
	}

}
