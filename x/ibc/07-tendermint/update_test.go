package tendermint_test

import (
	"bytes"
	"time"

	lite "github.com/tendermint/tendermint/lite2"
	tmtypes "github.com/tendermint/tendermint/types"

	tendermint "github.com/KiraCore/cosmos-sdk/x/ibc/07-tendermint"
	ibctmtypes "github.com/KiraCore/cosmos-sdk/x/ibc/07-tendermint/types"
	commitmenttypes "github.com/KiraCore/cosmos-sdk/x/ibc/23-commitment/types"
)

func (suite *TendermintTestSuite) TestCheckValidity() {
	var (
		clientState ibctmtypes.ClientState
		newHeader   ibctmtypes.Header
		currentTime time.Time
	)

	// Setup different validators and signers for testing different types of updates
	altPrivVal := tmtypes.NewMockPV()
	altPubKey, err := altPrivVal.GetPubKey()
	suite.Require().NoError(err)

	altVal := tmtypes.NewValidator(altPubKey, height)

	// Create bothValSet with both suite validator and altVal. Would be valid update
	bothValSet := tmtypes.NewValidatorSet(append(suite.valSet.Validators, altVal))
	// Create alternative validator set with only altVal, invalid update (too much change in valSet)
	altValSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{altVal})

	signers := []tmtypes.PrivValidator{suite.privVal}

	pubKey, err := suite.privVal.GetPubKey()
	suite.Require().NoError(err)

	// Create signer array and ensure it is in same order as bothValSet
	var bothSigners []tmtypes.PrivValidator
	if bytes.Compare(altPubKey.Address(), pubKey.Address()) == -1 {
		bothSigners = []tmtypes.PrivValidator{altPrivVal, suite.privVal}
	} else {
		bothSigners = []tmtypes.PrivValidator{suite.privVal, altPrivVal}
	}

	altSigners := []tmtypes.PrivValidator{altPrivVal}

	testCases := []struct {
		name    string
		setup   func()
		expPass bool
	}{
		{
			name: "successful update with next height and same validator set",
			setup: func() {
				clientState = ibctmtypes.NewClientState(chainID, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, suite.header, commitmenttypes.GetSDKSpecs())
				newHeader = ibctmtypes.CreateTestHeader(chainID, height+1, suite.headerTime, suite.valSet, signers)
				currentTime = suite.now
			},
			expPass: true,
		},
		{
			name: "successful update with future height and different validator set",
			setup: func() {
				clientState = ibctmtypes.NewClientState(chainID, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, suite.header, commitmenttypes.GetSDKSpecs())
				newHeader = ibctmtypes.CreateTestHeader(chainID, height+5, suite.headerTime, bothValSet, bothSigners)
				currentTime = suite.now
			},
			expPass: true,
		},
		{
			name: "unsuccessful update with next height: update header mismatches nextValSetHash",
			setup: func() {
				clientState = ibctmtypes.NewClientState(chainID, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, suite.header, commitmenttypes.GetSDKSpecs())
				newHeader = ibctmtypes.CreateTestHeader(chainID, height+1, suite.headerTime, bothValSet, bothSigners)
				currentTime = suite.now
			},
			expPass: false,
		},
		{
			name: "unsuccessful update with future height: too much change in validator set",
			setup: func() {
				clientState = ibctmtypes.NewClientState(chainID, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, suite.header, commitmenttypes.GetSDKSpecs())
				newHeader = ibctmtypes.CreateTestHeader(chainID, height+5, suite.headerTime, altValSet, altSigners)
				currentTime = suite.now
			},
			expPass: false,
		},
		{
			name: "unsuccessful update: trusting period has passed since last client timestamp",
			setup: func() {
				clientState = ibctmtypes.NewClientState(chainID, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, suite.header, commitmenttypes.GetSDKSpecs())
				newHeader = ibctmtypes.CreateTestHeader(chainID, height+1, suite.headerTime, suite.valSet, signers)
				// make current time pass trusting period from last timestamp on clientstate
				currentTime = suite.now.Add(ubdPeriod)
			},
			expPass: false,
		},
		{
			name: "unsuccessful update: header timestamp is past current timestamp",
			setup: func() {
				clientState = ibctmtypes.NewClientState(chainID, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, suite.header, commitmenttypes.GetSDKSpecs())
				newHeader = ibctmtypes.CreateTestHeader(chainID, height+1, suite.now.Add(time.Minute), suite.valSet, signers)
				currentTime = suite.now
			},
			expPass: false,
		},
		{
			name: "unsuccessful update: header timestamp is not past last client timestamp",
			setup: func() {
				clientState = ibctmtypes.NewClientState(chainID, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, suite.header, commitmenttypes.GetSDKSpecs())
				newHeader = ibctmtypes.CreateTestHeader(chainID, height+1, suite.clientTime, suite.valSet, signers)
				currentTime = suite.now
			},
			expPass: false,
		},
		{
			name: "header basic validation failed",
			setup: func() {
				clientState = ibctmtypes.NewClientState(chainID, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, suite.header, commitmenttypes.GetSDKSpecs())
				newHeader = ibctmtypes.CreateTestHeader(chainID, height+1, suite.headerTime, suite.valSet, signers)
				// cause new header to fail validatebasic by changing commit height to mismatch header height
				newHeader.SignedHeader.Commit.Height = height - 1
				currentTime = suite.now
			},
			expPass: false,
		},
		{
			name: "header height < latest client height",
			setup: func() {
				clientState = ibctmtypes.NewClientState(chainID, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, suite.header, commitmenttypes.GetSDKSpecs())
				// Make new header at height less than latest client state
				newHeader = ibctmtypes.CreateTestHeader(chainID, height-1, suite.headerTime, suite.valSet, signers)
				currentTime = suite.now
			},

			expPass: false,
		},
	}

	for i, tc := range testCases {
		tc := tc
		// setup test
		tc.setup()

		expectedConsensus := ibctmtypes.ConsensusState{
			Height:       uint64(newHeader.Height),
			Timestamp:    newHeader.Time,
			Root:         commitmenttypes.NewMerkleRoot(newHeader.AppHash),
			ValidatorSet: newHeader.ValidatorSet,
		}

		clientState, consensusState, err := tendermint.CheckValidityAndUpdateState(clientState, newHeader, currentTime)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
			suite.Require().Equal(newHeader.GetHeight(), clientState.GetLatestHeight(), "valid test case %d failed: %s", i, tc.name)
			suite.Require().Equal(expectedConsensus, consensusState, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
			suite.Require().Nil(clientState, "invalid test case %d passed: %s", i, tc.name)
			suite.Require().Nil(consensusState, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}
