package ibc_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/types"
)

var (
	clientTime = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
)

type ProposalHandlerTestSuite struct {
	suite.Suite

	cdc *codec.LegacyAmino
	ctx sdk.Context
	app *simapp.SimApp

	header         ibctmtypes.Header
	consensusState *ibctmtypes.ConsensusState
}

func (suite *ProposalHandlerTestSuite) SetupTest() {
	isCheckTx := false
	suite.app = simapp.Setup(isCheckTx)

	privVal := tmtypes.NewMockPV()
	pubKey, err := privVal.GetPubKey()
	suite.Require().NoError(err)

	val := tmtypes.NewValidator(pubKey, 10)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{val})

	suite.header = ibctmtypes.CreateTestHeader(chainID, height+1, height, clientTime.Add(time.Second*5), valSet, valSet, []tmtypes.PrivValidator{privVal})

	suite.cdc = suite.app.LegacyAmino()
	suite.ctx = suite.app.BaseApp.NewContext(isCheckTx, abci.Header{Time: clientTime})

	suite.consensusState = ibctmtypes.NewConsensusState(clientTime, commitmenttypes.NewMerkleRoot([]byte("hash")), height, valSet.Hash())

}

func TestProposalHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ProposalHandlerTestSuite))
}

func testClientUpdateProposal(clientID string, header ibctmtypes.Header) (*types.ClientUpdateProposal, error) {
	return types.NewClientUpdateProposal("Test", "description", clientID, header)
}

func (suite *ProposalHandlerTestSuite) TestClientUpdateProposalHandlerPassed2() {
	ibcKeeper := *suite.app.IBCKeeper
	clientkeeper := ibcKeeper.ClientKeeper

	clientState := ibctmtypes.NewClientState(chainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, latestTimestamp, commitmenttypes.GetSDKSpecs(), true, true)

	_, err := clientkeeper.CreateClient(suite.ctx, clientID, clientState, suite.consensusState)
	suite.Require().NoError(err)

	p, err := testClientUpdateProposal(clientID, suite.header)
	suite.Require().NoError(err)

	hdlr := ibc.NewClientUpdateProposalHandler(ibcKeeper)
	err = hdlr(suite.ctx, p)
	suite.Require().NoError(err)
}

func (suite *ProposalHandlerTestSuite) testClientState(allowGovernanceOverrideAfterExpire bool, isExpire bool, allowGovernanceOverrideAfterMisbehaviour bool, isFrozen bool) *ibctmtypes.ClientState {
	clientState := ibctmtypes.NewClientState(chainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod,
		maxClockDrift, height, time.Now(), commitmenttypes.GetSDKSpecs(),
		allowGovernanceOverrideAfterExpire, allowGovernanceOverrideAfterMisbehaviour)

	// overwrite client FrozenHeight based on the isFrozen flag
	if isFrozen {
		clientState.FrozenHeight = 15
		suite.Require().True(clientState.IsFrozen(), "expected client to be frozen")
	} else {
		clientState.FrozenHeight = 0
		suite.Require().False(clientState.IsFrozen(), "expected client to not be frozen")
	}

	// overwrite client LatestTimestamp based on the isExpire flag
	if isExpire {
		clientState.LatestTimestamp = clientTime
		suite.Require().True(clientState.Expired(), "expected client to be expired")
	} else {
		clientState.LatestTimestamp = time.Now()
		suite.Require().False(clientState.Expired(), "expected client not to be expired")
	}

	return clientState
}

func (suite *ProposalHandlerTestSuite) testStoredClientStatus(exptedHeight uint64, expectedIsFrozen bool, clientID string) {
	ibcKeeper := *suite.app.IBCKeeper
	clientkeeper := ibcKeeper.ClientKeeper
	clientStatus, ok := clientkeeper.GetClientState(suite.ctx, clientID)
	suite.Require().True(ok)
	suite.Require().Equal(exptedHeight, clientStatus.GetLatestHeight())
	suite.Require().Equal(expectedIsFrozen, clientStatus.IsFrozen())
}

func (suite *ProposalHandlerTestSuite) TestClientUpdateProposalHandler() {
	testCases := []struct {
		name        string
		clientState *ibctmtypes.ClientState
		expPass     bool
	}{
		// tests all combinations of the the following flags:
		// allowGovernanceOverrideAfterExpire, isExpire
		// allowGovernanceOverrideAfterMisbehaviour, isFrozen
		{"test 1", suite.testClientState(false, false, false, false), false},
		{"test 2", suite.testClientState(false, false, false, true), false},
		{"test 3", suite.testClientState(false, false, true, false), false},
		{"test 4", suite.testClientState(false, false, true, true), true},
		{"test 5", suite.testClientState(false, true, false, false), false},
		{"test 6", suite.testClientState(false, true, false, true), false},
		{"test 7", suite.testClientState(false, true, true, false), false},
		{"test 8", suite.testClientState(false, true, true, true), true},
		{"test 9", suite.testClientState(true, false, false, false), false},
		{"test 10", suite.testClientState(true, false, false, true), false},
		{"test 11", suite.testClientState(true, false, true, false), false},
		{"test 12", suite.testClientState(true, false, true, true), true},
		{"test 13", suite.testClientState(true, true, false, false), true},
		{"test 14", suite.testClientState(true, true, false, true), true},
		{"test 15", suite.testClientState(true, true, true, false), true},
		{"test 16", suite.testClientState(true, true, true, true), true},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			ibcKeeper := *suite.app.IBCKeeper
			clientkeeper := ibcKeeper.ClientKeeper

			// create test client
			_, err := clientkeeper.CreateClient(suite.ctx, clientID, tc.clientState, suite.consensusState)
			suite.Require().NoError(err)

			// check stored clientStatus before update
			suite.testStoredClientStatus(10, tc.clientState.IsFrozen(), clientID)

			// create proposal
			p, err := testClientUpdateProposal(clientID, suite.header)
			suite.Require().NoError(err)

			// handle proposal
			hdlr := ibc.NewClientUpdateProposalHandler(ibcKeeper)
			err = hdlr(suite.ctx, p)

			if tc.expPass {
				suite.Require().NoError(err)

				isFrozen := tc.clientState.IsFrozen()
				if tc.clientState.IsFrozen() && tc.clientState.AllowGovernanceOverrideAfterMisbehaviour {
					isFrozen = false
				}
				// check stored clientStatus after update
				suite.testStoredClientStatus(11, isFrozen, clientID)

			} else {
				suite.Require().Error(err, err)
			}
		})
	}

}
