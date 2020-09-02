package types_test

/*
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
)

const (
	chainID  = "chainID"
	clientID = "clientidone"

	height = 10

	trustingPeriod time.Duration = time.Hour * 24 * 7 * 2
	ubdPeriod      time.Duration = time.Hour * 24 * 7 * 3
	maxClockDrift  time.Duration = time.Second * 10
)

var latestTimestamp = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

type ProposalHandlerTestSuite struct {
	suite.Suite

	cdc *codec.LegacyAmino
	ctx sdk.Context
	app *simapp.SimApp

	header         *ibctmtypes.Header
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

	suite.header = ibctmtypes.CreateTestHeader(chainID, height+1, height, latestTimestamp.Add(time.Second*5), valSet, valSet, []tmtypes.PrivValidator{privVal})

	suite.cdc = suite.app.LegacyAmino()
	suite.ctx = suite.app.BaseApp.NewContext(isCheckTx, tmproto.Header{Time: latestTimestamp.Add(time.Second * 10)})
	suite.consensusState = ibctmtypes.NewConsensusState(latestTimestamp, commitmenttypes.NewMerkleRoot([]byte("hash")), height, valSet.Hash())

}

func TestProposalHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ProposalHandlerTestSuite))
}

func testClientUpdateProposal(clientID string, header *ibctmtypes.Header) (*clienttypes.ClientUpdateProposal, error) {
	return clienttypes.NewClientUpdateProposal("Test", "description", clientID, header)
}

func (suite *ProposalHandlerTestSuite) testClientState(allowGovernanceOverrideAfterExpire bool, latestTimestamp time.Time, allowGovernanceOverrideAfterMisbehaviour bool, frozenHeight uint64) *ibctmtypes.ClientState {
	clientState := ibctmtypes.NewClientState(chainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod,
		maxClockDrift, height, latestTimestamp, commitmenttypes.GetSDKSpecs(),
		allowGovernanceOverrideAfterExpire, allowGovernanceOverrideAfterMisbehaviour)

	if frozenHeight > 0 {
		clientState.FrozenHeight = frozenHeight
	}

	return clientState
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

	testCases := []struct {
		name            string
		ctx             sdk.Context
		clientIsExpired bool
		clientIsFrozen  bool
		clientState     *ibctmtypes.ClientState
		expPass         bool
	}{
		// abbreviation:
		// OAE := allowGovernanceOverrideAfterExpire
		// OAM := allowGovernanceOverrideAfterMisbehaviour
		{
			"Test1 should fail for clientStatus with OAE=false, Expired=false, OAM=false, Frozen=false",
			suite.ctx, false, false, suite.testClientState(false, latestTimestamp, false, 0), false,
		},
		{
			"Test2 should fail for clientStatus with OAE=false, Expired=false, OAM=false, Frozen=True",
			suite.ctx, false, true, suite.testClientState(false, latestTimestamp, false, 2), false,
		},
		{
			"Test3 should fail for clientStatus with OAE=false, Expired=false, OAM=true, Frozen=false",
			suite.ctx, false, false, suite.testClientState(false, latestTimestamp, true, 0), false,
		},
		{
			"Test4 should pass for clientStatus with OAE=false, Expired=false, OAM=true, Frozen=true",
			suite.ctx, false, true, suite.testClientState(false, latestTimestamp, true, 2), true,
		},
		{
			"Test5 should fail for clientStatus with OAE=false, Expired=true, OAM=false, Frozen=false",
			expiredCtx, true, false, suite.testClientState(false, latestTimestamp, false, 0), false,
		},
		{
			"Test6 should fail for clientStatus with OAE=false, Expired=true, OAM=false, Frozen=true",
			expiredCtx, true, true, suite.testClientState(false, latestTimestamp, false, 2), false,
		},
		{
			"Test7 should fail for clientStatus with OAE=false, Expired=true, OAM=true, Frozen=false",
			expiredCtx, true, false, suite.testClientState(false, latestTimestamp, true, 0), false,
		},
		{
			// For this test, the client update proposal will pass but the client update will
			// fail because the new header, (suite.header), is expired and the override flag is
			// set to be false.
			"Test8 should pass for clientStatus with OAE=false, Expired=true, OAM=true, Frozen=true",
			expiredCtx, true, true, suite.testClientState(false, latestTimestamp, true, 2), false,
		},
		{
			"Test10 should fail for clientStatus with OAE=true, Expired=false, OAM=false, Frozen=false",
			suite.ctx, false, false, suite.testClientState(true, latestTimestamp, false, 0), false,
		},
		{
			"Test11 should fail for clientStatus with OAE=true, Expired=false, OAM=false, Frozen=true",
			suite.ctx, false, true, suite.testClientState(true, latestTimestamp, false, 2), false,
		},
		{
			"Test12 should fail for clientStatus with OAE=true, Expired=false, OAM=false, Frozen=true",
			suite.ctx, false, false, suite.testClientState(true, latestTimestamp, true, 0), false,
		},
		{
			"Test13 should pass for clientStatus with OAE=true, Expired=false, OAM=true, Frozen=true",
			suite.ctx, false, true, suite.testClientState(true, latestTimestamp, true, 2), true,
		},
		{
			"Test14 should pass for clientStatus with OAE=true, Expired=true, OAM=false, Frozen=false",
			expiredCtx, true, false, suite.testClientState(true, latestTimestamp, false, 0), true,
		},
		{
			// For this test, the client update proposal will pass but but the client update will
			// fail (even though the override flag is set to be true) because the client is still frozen
			// and OAM = false
			"Test15 should fail for clientStatus with OAE=true, Expired=true, OAM=false, Frozen=true",
			expiredCtx, true, true, suite.testClientState(true, latestTimestamp, false, 2), false,
		},
		{
			"Test16 should pass for clientStatus with OAE=true, Expired=true, OAM=true, Frozen=false",
			expiredCtx, true, false, suite.testClientState(true, latestTimestamp, true, 0), true,
		},
		{
			"Test17 should pass for clientStatus with OAE=true, Expired=true, OAM=true, Frozen=true",
			expiredCtx, true, true, suite.testClientState(true, latestTimestamp, true, 2), true,
		},
	}

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
*/
