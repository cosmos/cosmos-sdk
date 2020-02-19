package tendermint_test

import (
	"bytes"

	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	tmtypes "github.com/tendermint/tendermint/types"
)

func (suite *TendermintTestSuite) TestCheckValidity() {
	var (
		clientState ibctmtypes.ClientState
		newHeader   ibctmtypes.Header
	)

	// Setup different validators and signers for testing different types of updates
	altPrivVal := tmtypes.NewMockPV()
	altVal := tmtypes.NewValidator(altPrivVal.GetPubKey(), height)

	// Create bothValSet with both suite validator and altVal. Would be valid update
	bothValSet := tmtypes.NewValidatorSet(append(suite.valSet.Validators, altVal))
	// Create alternative validator set with only altVal, invalid update (too much change in valSet)
	altValSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{altVal})

	signers := []tmtypes.PrivValidator{suite.privVal}
	// Create signer array and ensure it is in same order as bothValSet
	var bothSigners []tmtypes.PrivValidator
	if bytes.Compare(altPrivVal.GetPubKey().Address(), suite.privVal.GetPubKey().Address()) == -1 {
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
				clientState = ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, suite.header)
				newHeader = ibctmtypes.CreateTestHeader(chainID, height+1, suite.headerTime, suite.valSet, suite.valSet, signers)
			},
			expPass: true,
		},
		{
			name: "successful update with next height and different validator set",
			setup: func() {
				oldHeader := ibctmtypes.CreateTestHeader(chainID, height, suite.clientTime, suite.valSet, bothValSet, bothSigners)
				clientState = ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, oldHeader)
				newHeader = ibctmtypes.CreateTestHeader(chainID, height+1, suite.headerTime, bothValSet, bothValSet, bothSigners)
			},
			expPass: true,
		},
		{
			name: "successful update with future height and same validator set",
			setup: func() {
				clientState = ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, suite.header)
				newHeader = ibctmtypes.CreateTestHeader(chainID, height+5, suite.headerTime, bothValSet, bothValSet, bothSigners)
			},
			expPass: true,
		},
		{
			name: "unsuccessful update with next height: update header mismatches nextValSetHash",
			setup: func() {
				clientState = ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, suite.header)
				newHeader = ibctmtypes.CreateTestHeader(chainID, height+1, suite.headerTime, bothValSet, bothValSet, bothSigners)
			},
			expPass: false,
		},
		{
			name: "unsuccessful update with future height: too much change in validator set",
			setup: func() {
				clientState = ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, suite.header)
				newHeader = ibctmtypes.CreateTestHeader(chainID, height+5, suite.headerTime, altValSet, altValSet, altSigners)
			},
			expPass: true,
		},

		{
			name: "header basic validation failed",
			setup: func() {
				clientState = ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, suite.header)
				newHeader = ibctmtypes.CreateTestHeader(chainID, height+1, suite.headerTime, suite.valSet, suite.valSet, signers)
				// cause new header to fail validatebasic by changing commit height to mismatch header height
				newHeader.SignedHeader.Commit.Height = height - 1
			},
			expPass: false,
		},
		{
			name: "header height < latest client height",
			setup: func() {
				clientState = ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, suite.header)
				// Make new header at height less than latest client state
				newHeader = ibctmtypes.CreateTestHeader(chainID, height-1, suite.headerTime, suite.valSet, suite.valSet, signers)
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
			Timestamp:    suite.headerTime,
			Root:         commitment.NewRoot(suite.header.AppHash),
			ValidatorSet: suite.header.ValidatorSet,
		}

		clientState, consensusState, err := tendermint.CheckValidityAndUpdateState(clientState, newHeader, suite.now)

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
