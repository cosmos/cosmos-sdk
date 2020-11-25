package types_test

import (
	"time"

	tmtypes "github.com/tendermint/tendermint/types"

	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/23-commitment/types"
	types "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
	ibctestingmock "github.com/cosmos/cosmos-sdk/x/ibc/testing/mock"
)

func (suite *TendermintTestSuite) TestCheckHeaderAndUpdateState() {
	var (
		clientState     *types.ClientState
		consensusState  *types.ConsensusState
		consStateHeight clienttypes.Height
		newHeader       *types.Header
		currentTime     time.Time
	)

	// Setup different validators and signers for testing different types of updates
	altPrivVal := ibctestingmock.NewPV()
	altPubKey, err := altPrivVal.GetPubKey()
	suite.Require().NoError(err)

	revisionHeight := int64(height.RevisionHeight)

	// create modified heights to use for test-cases
	heightPlus1 := clienttypes.NewHeight(height.RevisionNumber, height.RevisionHeight+1)
	heightMinus1 := clienttypes.NewHeight(height.RevisionNumber, height.RevisionHeight-1)
	heightMinus3 := clienttypes.NewHeight(height.RevisionNumber, height.RevisionHeight-3)
	heightPlus5 := clienttypes.NewHeight(height.RevisionNumber, height.RevisionHeight+5)

	altVal := tmtypes.NewValidator(altPubKey, revisionHeight)

	// Create bothValSet with both suite validator and altVal. Would be valid update
	bothValSet := tmtypes.NewValidatorSet(append(suite.valSet.Validators, altVal))
	// Create alternative validator set with only altVal, invalid update (too much change in valSet)
	altValSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{altVal})

	signers := []tmtypes.PrivValidator{suite.privVal}

	// Create signer array and ensure it is in same order as bothValSet
	_, suiteVal := suite.valSet.GetByIndex(0)
	bothSigners := ibctesting.CreateSortedSignerArray(altPrivVal, suite.privVal, altVal, suiteVal)

	altSigners := []tmtypes.PrivValidator{altPrivVal}

	testCases := []struct {
		name    string
		setup   func()
		expPass bool
	}{
		{
			name: "successful update with next height and same validator set",
			setup: func() {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), suite.valsHash)
				newHeader = suite.chainA.CreateTMClientHeader(chainID, int64(heightPlus1.RevisionHeight), height, suite.headerTime, suite.valSet, suite.valSet, signers)
				currentTime = suite.now
			},
			expPass: true,
		},
		{
			name: "successful update with future height and different validator set",
			setup: func() {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), suite.valsHash)
				newHeader = suite.chainA.CreateTMClientHeader(chainID, int64(heightPlus5.RevisionHeight), height, suite.headerTime, bothValSet, suite.valSet, bothSigners)
				currentTime = suite.now
			},
			expPass: true,
		},
		{
			name: "successful update with next height and different validator set",
			setup: func() {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), bothValSet.Hash())
				newHeader = suite.chainA.CreateTMClientHeader(chainID, int64(heightPlus1.RevisionHeight), height, suite.headerTime, bothValSet, bothValSet, bothSigners)
				currentTime = suite.now
			},
			expPass: true,
		},
		{
			name: "successful update for a previous height",
			setup: func() {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), suite.valsHash)
				consStateHeight = heightMinus3
				newHeader = suite.chainA.CreateTMClientHeader(chainID, int64(heightMinus1.RevisionHeight), heightMinus3, suite.headerTime, bothValSet, suite.valSet, bothSigners)
				currentTime = suite.now
			},
			expPass: true,
		},
		{
			name: "successful update for a previous revision",
			setup: func() {
				clientState = types.NewClientState(chainIDRevision1, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), suite.valsHash)
				newHeader = suite.chainA.CreateTMClientHeader(chainIDRevision0, int64(height.RevisionHeight), heightMinus3, suite.headerTime, bothValSet, suite.valSet, bothSigners)
				currentTime = suite.now
			},
			expPass: true,
		},
		{
			name: "unsuccessful update with incorrect header chain-id",
			setup: func() {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), suite.valsHash)
				newHeader = suite.chainA.CreateTMClientHeader("ethermint", int64(heightPlus1.RevisionHeight), height, suite.headerTime, suite.valSet, suite.valSet, signers)
				currentTime = suite.now
			},
			expPass: false,
		},
		{
			name: "unsuccessful update to a future revision",
			setup: func() {
				clientState = types.NewClientState(chainIDRevision0, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), suite.valsHash)
				newHeader = suite.chainA.CreateTMClientHeader(chainIDRevision1, 1, height, suite.headerTime, suite.valSet, suite.valSet, signers)
				currentTime = suite.now
			},
			expPass: false,
		},
		{
			name: "unsuccessful update: header height revision and trusted height revision mismatch",
			setup: func() {
				clientState = types.NewClientState(chainIDRevision1, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, clienttypes.NewHeight(1, 1), commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), suite.valsHash)
				newHeader = suite.chainA.CreateTMClientHeader(chainIDRevision1, 3, height, suite.headerTime, suite.valSet, suite.valSet, signers)
				currentTime = suite.now
			},
			expPass: false,
		},
		{
			name: "unsuccessful update with next height: update header mismatches nextValSetHash",
			setup: func() {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), suite.valsHash)
				newHeader = suite.chainA.CreateTMClientHeader(chainID, int64(heightPlus1.RevisionHeight), height, suite.headerTime, bothValSet, suite.valSet, bothSigners)
				currentTime = suite.now
			},
			expPass: false,
		},
		{
			name: "unsuccessful update with next height: update header mismatches different nextValSetHash",
			setup: func() {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), bothValSet.Hash())
				newHeader = suite.chainA.CreateTMClientHeader(chainID, int64(heightPlus1.RevisionHeight), height, suite.headerTime, suite.valSet, bothValSet, signers)
				currentTime = suite.now
			},
			expPass: false,
		},
		{
			name: "unsuccessful update with future height: too much change in validator set",
			setup: func() {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), suite.valsHash)
				newHeader = suite.chainA.CreateTMClientHeader(chainID, int64(heightPlus5.RevisionHeight), height, suite.headerTime, altValSet, suite.valSet, altSigners)
				currentTime = suite.now
			},
			expPass: false,
		},
		{
			name: "unsuccessful updates, passed in incorrect trusted validators for given consensus state",
			setup: func() {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), suite.valsHash)
				newHeader = suite.chainA.CreateTMClientHeader(chainID, int64(heightPlus5.RevisionHeight), height, suite.headerTime, bothValSet, bothValSet, bothSigners)
				currentTime = suite.now
			},
			expPass: false,
		},
		{
			name: "unsuccessful update: trusting period has passed since last client timestamp",
			setup: func() {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), suite.valsHash)
				newHeader = suite.chainA.CreateTMClientHeader(chainID, int64(heightPlus1.RevisionHeight), height, suite.headerTime, suite.valSet, suite.valSet, signers)
				// make current time pass trusting period from last timestamp on clientstate
				currentTime = suite.now.Add(trustingPeriod)
			},
			expPass: false,
		},
		{
			name: "unsuccessful update: header timestamp is past current timestamp",
			setup: func() {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), suite.valsHash)
				newHeader = suite.chainA.CreateTMClientHeader(chainID, int64(heightPlus1.RevisionHeight), height, suite.now.Add(time.Minute), suite.valSet, suite.valSet, signers)
				currentTime = suite.now
			},
			expPass: false,
		},
		{
			name: "unsuccessful update: header timestamp is not past last client timestamp",
			setup: func() {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), suite.valsHash)
				newHeader = suite.chainA.CreateTMClientHeader(chainID, int64(heightPlus1.RevisionHeight), height, suite.clientTime, suite.valSet, suite.valSet, signers)
				currentTime = suite.now
			},
			expPass: false,
		},
		{
			name: "header basic validation failed",
			setup: func() {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), suite.valsHash)
				newHeader = suite.chainA.CreateTMClientHeader(chainID, int64(heightPlus1.RevisionHeight), height, suite.headerTime, suite.valSet, suite.valSet, signers)
				// cause new header to fail validatebasic by changing commit height to mismatch header height
				newHeader.SignedHeader.Commit.Height = revisionHeight - 1
				currentTime = suite.now
			},
			expPass: false,
		},
		{
			name: "header height < consensus height",
			setup: func() {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, clienttypes.NewHeight(height.RevisionNumber, heightPlus5.RevisionHeight), commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), suite.valsHash)
				// Make new header at height less than latest client state
				newHeader = suite.chainA.CreateTMClientHeader(chainID, int64(heightMinus1.RevisionHeight), height, suite.headerTime, suite.valSet, suite.valSet, signers)
				currentTime = suite.now
			},
			expPass: false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		consStateHeight = height // must be explicitly changed
		// setup test
		tc.setup()

		// Set current timestamp in context
		ctx := suite.chainA.GetContext().WithBlockTime(currentTime)

		// Set trusted consensus state in client store
		suite.chainA.App.IBCKeeper.ClientKeeper.SetClientConsensusState(ctx, clientID, consStateHeight, consensusState)

		height := newHeader.GetHeight()
		expectedConsensus := &types.ConsensusState{
			Timestamp:          newHeader.GetTime(),
			Root:               commitmenttypes.NewMerkleRoot(newHeader.Header.GetAppHash()),
			NextValidatorsHash: newHeader.Header.NextValidatorsHash,
		}

		newClientState, consensusState, err := clientState.CheckHeaderAndUpdateState(
			ctx,
			suite.cdc,
			suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(suite.chainA.GetContext(), clientID), // pass in clientID prefixed clientStore
			newHeader,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)

			// Determine if clientState should be updated or not
			// TODO: check the entire Height struct once GetLatestHeight returns clienttypes.Height
			if height.GT(clientState.LatestHeight) {
				// Header Height is greater than clientState latest Height, clientState should be updated with header.GetHeight()
				suite.Require().Equal(height, newClientState.GetLatestHeight(), "clientstate height did not update")
			} else {
				// Update will add past consensus state, clientState should not be updated at all
				suite.Require().Equal(clientState.LatestHeight, newClientState.GetLatestHeight(), "client state height updated for past header")
			}

			suite.Require().Equal(expectedConsensus, consensusState, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
			suite.Require().Nil(newClientState, "invalid test case %d passed: %s", i, tc.name)
			suite.Require().Nil(consensusState, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}
