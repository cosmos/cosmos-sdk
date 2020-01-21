package tendermint_test

import (
	"bytes"

	"github.com/tendermint/tendermint/crypto/tmhash"
	tmtypes "github.com/tendermint/tendermint/types"

	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint"
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
		clientState    tendermint.ClientState
		consensusState tendermint.ConsensusState
		evidence       tendermint.Evidence
		height         uint64
		expPass        bool
	}{
		{
			"valid misbehavior evidence",
			tendermint.ClientState{ID: chainID, LatestHeight: height, FrozenHeight: 0},
			tendermint.ConsensusState{Root: commitment.NewRoot(tmhash.Sum([]byte("app_hash"))), ValidatorSetHash: bothValSet.Hash()},
			tendermint.Evidence{
				Header1:          tendermint.CreateTestHeader(chainID, height, bothValSet, suite.valSet, bothSigners),
				Header2:          tendermint.CreateTestHeader(chainID, height, bothValSet, bothValSet, bothSigners),
				FromValidatorSet: bothValSet,
				ChainID:          chainID,
				ClientID:         chainID,
			},
			height,
			true,
		},
		{
			"height doesn't match provided evidence",
			tendermint.ClientState{ID: chainID, LatestHeight: height, FrozenHeight: 0},
			tendermint.ConsensusState{Root: commitment.NewRoot(tmhash.Sum([]byte("app_hash"))), ValidatorSetHash: bothValSet.Hash()},
			tendermint.Evidence{
				Header1:          tendermint.CreateTestHeader(chainID, height, bothValSet, suite.valSet, bothSigners),
				Header2:          tendermint.CreateTestHeader(chainID, height, bothValSet, bothValSet, bothSigners),
				FromValidatorSet: bothValSet,
				ChainID:          chainID,
				ClientID:         chainID,
			},
			0,
			false,
		},
		{
			"consensus state's valset hash different from evidence",
			tendermint.ClientState{ID: chainID, LatestHeight: height, FrozenHeight: 0},
			tendermint.ConsensusState{Root: commitment.NewRoot(tmhash.Sum([]byte("app_hash"))), ValidatorSetHash: suite.valSet.Hash()},
			tendermint.Evidence{
				Header1:          tendermint.CreateTestHeader(chainID, height, bothValSet, suite.valSet, bothSigners),
				Header2:          tendermint.CreateTestHeader(chainID, height, bothValSet, bothValSet, bothSigners),
				FromValidatorSet: bothValSet,
				ChainID:          chainID,
				ClientID:         chainID,
			},
			height,
			false,
		},
		{
			"first valset has too much change",
			tendermint.ClientState{ID: chainID, LatestHeight: height, FrozenHeight: 0},
			tendermint.ConsensusState{Root: commitment.NewRoot(tmhash.Sum([]byte("app_hash"))), ValidatorSetHash: bothValSet.Hash()},
			tendermint.Evidence{
				Header1:          tendermint.CreateTestHeader(chainID, height, altValSet, bothValSet, altSigners),
				Header2:          tendermint.CreateTestHeader(chainID, height, bothValSet, bothValSet, bothSigners),
				FromValidatorSet: bothValSet,
				ChainID:          chainID,
				ClientID:         chainID,
			},
			height,
			false,
		},
		{
			"second valset has too much change",
			tendermint.ClientState{ID: chainID, LatestHeight: height, FrozenHeight: 0},
			tendermint.ConsensusState{Root: commitment.NewRoot(tmhash.Sum([]byte("app_hash"))), ValidatorSetHash: bothValSet.Hash()},
			tendermint.Evidence{
				Header1:          tendermint.CreateTestHeader(chainID, height, bothValSet, bothValSet, bothSigners),
				Header2:          tendermint.CreateTestHeader(chainID, height, altValSet, bothValSet, altSigners),
				FromValidatorSet: bothValSet,
				ChainID:          chainID,
				ClientID:         chainID,
			},
			height,
			false,
		},
		{
			"both valsets have too much change",
			tendermint.ClientState{ID: chainID, LatestHeight: height, FrozenHeight: 0},
			tendermint.ConsensusState{Root: commitment.NewRoot(tmhash.Sum([]byte("app_hash"))), ValidatorSetHash: bothValSet.Hash()},
			tendermint.Evidence{
				Header1:          tendermint.CreateTestHeader(chainID, height, altValSet, altValSet, altSigners),
				Header2:          tendermint.CreateTestHeader(chainID, height, altValSet, bothValSet, altSigners),
				FromValidatorSet: bothValSet,
				ChainID:          chainID,
				ClientID:         chainID,
			},
			height,
			false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		clientState, err := tendermint.CheckMisbehaviourAndUpdateState(tc.clientState, tc.consensusState, tc.evidence, tc.height)

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
