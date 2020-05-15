package tendermint_test

import (
	"bytes"
	"time"

	"github.com/tendermint/tendermint/crypto/tmhash"
	lite "github.com/tendermint/tendermint/lite2"
	tmtypes "github.com/tendermint/tendermint/types"

	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

func (suite *TendermintTestSuite) TestCheckMisbehaviour() {
	altPrivVal := tmtypes.NewMockPV()
	altPubKey, err := altPrivVal.GetPubKey()
	suite.Require().NoError(err)

	altVal := tmtypes.NewValidator(altPubKey, 4)

	// Create bothValSet with both suite validator and altVal
	bothValSet := tmtypes.NewValidatorSet(append(suite.valSet.Validators, altVal))
	// Create alternative validator set with only altVal
	altValSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{altVal})

	// Create signer array and ensure it is in same order as bothValSet
	var bothSigners []tmtypes.PrivValidator

	pubKey, err := suite.privVal.GetPubKey()
	suite.Require().NoError(err)

	if bytes.Compare(altPubKey.Address(), pubKey.Address()) == -1 {
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
			ibctmtypes.NewClientState(chainID, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, suite.header),
			ibctmtypes.ConsensusState{Timestamp: suite.now, Root: commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), ValidatorSet: bothValSet},
			ibctmtypes.Evidence{
				Header1:  ibctmtypes.CreateTestHeader(chainID, height, suite.now, bothValSet, bothSigners),
				Header2:  ibctmtypes.CreateTestHeader(chainID, height, suite.now.Add(time.Minute), bothValSet, bothSigners),
				ChainID:  chainID,
				ClientID: chainID,
			},
			height,
			true,
		},
		{
			"valid misbehavior at height greater than last consensusState",
			ibctmtypes.NewClientState(chainID, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, suite.header),
			ibctmtypes.ConsensusState{Timestamp: suite.now, Height: height - 1, Root: commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), ValidatorSet: bothValSet},
			ibctmtypes.Evidence{
				Header1:  ibctmtypes.CreateTestHeader(chainID, height, suite.now, bothValSet, bothSigners),
				Header2:  ibctmtypes.CreateTestHeader(chainID, height, suite.now.Add(time.Minute), bothValSet, bothSigners),
				ChainID:  chainID,
				ClientID: chainID,
			},
			height - 1,
			true,
		},
		{
			"consensus state's valset hash different from evidence should still pass",
			ibctmtypes.NewClientState(chainID, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, suite.header),
			ibctmtypes.ConsensusState{Timestamp: suite.now, Height: height - 1, Root: commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), ValidatorSet: suite.valSet},
			ibctmtypes.Evidence{
				Header1:  ibctmtypes.CreateTestHeader(chainID, height, suite.now, bothValSet, bothSigners),
				Header2:  ibctmtypes.CreateTestHeader(chainID, height, suite.now.Add(time.Minute), bothValSet, bothSigners),
				ChainID:  chainID,
				ClientID: chainID,
			},
			height - 1,
			true,
		},
		{
			"first valset has too much change",
			ibctmtypes.NewClientState(chainID, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, suite.header),
			ibctmtypes.ConsensusState{Timestamp: suite.now, Root: commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), ValidatorSet: bothValSet},
			ibctmtypes.Evidence{
				Header1:  ibctmtypes.CreateTestHeader(chainID, height, suite.now, altValSet, altSigners),
				Header2:  ibctmtypes.CreateTestHeader(chainID, height, suite.now.Add(time.Minute), bothValSet, bothSigners),
				ChainID:  chainID,
				ClientID: chainID,
			},
			height,
			false,
		},
		{
			"second valset has too much change",
			ibctmtypes.NewClientState(chainID, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, suite.header),
			ibctmtypes.ConsensusState{Timestamp: suite.now, Root: commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), ValidatorSet: bothValSet},
			ibctmtypes.Evidence{
				Header1:  ibctmtypes.CreateTestHeader(chainID, height, suite.now, bothValSet, bothSigners),
				Header2:  ibctmtypes.CreateTestHeader(chainID, height, suite.now.Add(time.Minute), altValSet, altSigners),
				ChainID:  chainID,
				ClientID: chainID,
			},
			height,
			false,
		},
		{
			"both valsets have too much change",
			ibctmtypes.NewClientState(chainID, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, suite.header),
			ibctmtypes.ConsensusState{Timestamp: suite.now, Root: commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), ValidatorSet: bothValSet},
			ibctmtypes.Evidence{
				Header1:  ibctmtypes.CreateTestHeader(chainID, height, suite.now, altValSet, altSigners),
				Header2:  ibctmtypes.CreateTestHeader(chainID, height, suite.now.Add(time.Minute), altValSet, altSigners),
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
