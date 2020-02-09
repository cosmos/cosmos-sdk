package tendermint_test

import (
	"bytes"

	"github.com/tendermint/tendermint/crypto/tmhash"
	tmtypes "github.com/tendermint/tendermint/types"

	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

func (suite *TendermintTestSuite) TestCheckMisbehaviour() {
	altPrivVal := tmtypes.NewMockPV()
	altVal := tmtypes.NewValidator(altPrivVal.GetPubKey(), 4)

	// Create bothValSet with both suite validator and altVal
	bothValSet := tmtypes.NewValidatorSet(append(suite.valSet.Validators, altVal))
	// Create alternative validator set with only altVal
	altValSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{altVal})

	// Create signer array and ensure it is in same order as bothValSet
	var bothSigners []tmtypes.PrivValidator
	if bytes.Compare(altPrivVal.GetPubKey().Address(), suite.privVal.GetPubKey().Address()) == -1 {
		bothSigners = []tmtypes.PrivValidator{altPrivVal, suite.privVal}
	} else {
		bothSigners = []tmtypes.PrivValidator{suite.privVal, altPrivVal}
	}

	altSigners := []tmtypes.PrivValidator{altPrivVal}

	testCases := []struct {
		name           string
		clientState    ibctmtypes.ClientState
		consensusState ibctmtypes.ConsensusState
		evidence       ibctmtypes.Evidence
		height         uint64
		expPass        bool
	}{
		{
			"valid misbehavior evidence",
			ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, height, suite.now),
			ibctmtypes.ConsensusState{Timestamp: suite.now, Root: commitment.NewRoot(tmhash.Sum([]byte("app_hash"))), ValidatorSet: bothValSet},
			ibctmtypes.Evidence{
				Header1:  ibctmtypes.CreateTestHeader(chainID, height, suite.now, bothValSet, suite.valSet, bothSigners),
				Header2:  ibctmtypes.CreateTestHeader(chainID, height, suite.now, bothValSet, bothValSet, bothSigners),
				ChainID:  chainID,
				ClientID: chainID,
			},
			height,
			true,
		},
		{
			"height doesn't match provided evidence",
			ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, height, suite.now),
			ibctmtypes.ConsensusState{Timestamp: suite.now, Root: commitment.NewRoot(tmhash.Sum([]byte("app_hash"))), ValidatorSet: bothValSet},
			ibctmtypes.Evidence{
				Header1:  ibctmtypes.CreateTestHeader(chainID, height, suite.now, bothValSet, suite.valSet, bothSigners),
				Header2:  ibctmtypes.CreateTestHeader(chainID, height, suite.now, bothValSet, bothValSet, bothSigners),
				ChainID:  chainID,
				ClientID: chainID,
			},
			0,
			false,
		},
		// {
		// 	"consensus state's valset hash different from evidence",
		// 	ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, height, suite.now),
		// 	ibctmtypes.ConsensusState{Timestamp: suite.now, Root: commitment.NewRoot(tmhash.Sum([]byte("app_hash"))), ValidatorSet: suite.valSet},
		// 	ibctmtypes.Evidence{
		// 		Header1:  ibctmtypes.CreateTestHeader(chainID, height, suite.now, bothValSet, suite.valSet, bothSigners),
		// 		Header2:  ibctmtypes.CreateTestHeader(chainID, height, suite.now, bothValSet, bothValSet, bothSigners),
		// 		ChainID:  chainID,
		// 		ClientID: chainID,
		// 	},
		// 	height,
		// 	false,
		// },
		{
			"first valset has too much change",
			ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, height, suite.now),
			ibctmtypes.ConsensusState{Timestamp: suite.now, Root: commitment.NewRoot(tmhash.Sum([]byte("app_hash"))), ValidatorSet: bothValSet},
			ibctmtypes.Evidence{
				Header1:  ibctmtypes.CreateTestHeader(chainID, height, suite.now, altValSet, bothValSet, altSigners),
				Header2:  ibctmtypes.CreateTestHeader(chainID, height, suite.now, bothValSet, bothValSet, bothSigners),
				ChainID:  chainID,
				ClientID: chainID,
			},
			height,
			false,
		},
		{
			"second valset has too much change",
			ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, height, suite.now),
			ibctmtypes.ConsensusState{Timestamp: suite.now, Root: commitment.NewRoot(tmhash.Sum([]byte("app_hash"))), ValidatorSet: bothValSet},
			ibctmtypes.Evidence{
				Header1:  ibctmtypes.CreateTestHeader(chainID, height, suite.now, bothValSet, bothValSet, bothSigners),
				Header2:  ibctmtypes.CreateTestHeader(chainID, height, suite.now, altValSet, bothValSet, altSigners),
				ChainID:  chainID,
				ClientID: chainID,
			},
			height,
			false,
		},
		{
			"both valsets have too much change",
			ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, height, suite.now),
			ibctmtypes.ConsensusState{Timestamp: suite.now, Root: commitment.NewRoot(tmhash.Sum([]byte("app_hash"))), ValidatorSet: bothValSet},
			ibctmtypes.Evidence{
				Header1:  ibctmtypes.CreateTestHeader(chainID, height, suite.now, altValSet, altValSet, altSigners),
				Header2:  ibctmtypes.CreateTestHeader(chainID, height, suite.now, altValSet, bothValSet, altSigners),
				ChainID:  chainID,
				ClientID: chainID,
			},
			height,
			false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		clientState, err := tendermint.CheckMisbehaviourAndUpdateState(tc.clientState, tc.consensusState, tc.evidence, tc.height, suite.now)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
			suite.Require().NotNil(clientState, "valid test case %d failed: %s", i, tc.name)
			suite.Require().True(clientState.IsFrozen(), "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
			suite.Require().Nil(clientState, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}
