package types_test

import (
	"bytes"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint"
	"github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

func (suite *TendermintTestSuite) TestCheckMisbehaviourAndUpdateState() {
	altPrivVal := tmtypes.NewMockPV()
	altPubKey, err := altPrivVal.GetPubKey()
	suite.Require().NoError(err)

	altVal := tmtypes.NewValidator(altPubKey, 4)

	// Create bothValSet with both suite validator and altVal
	bothValSet := tmtypes.NewValidatorSet(append(suite.valSet.Validators, altVal))
	bothValsHash := bothValSet.Hash()
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
		name            string
		clientState     clientexported.ClientState
		consensusState1 clientexported.ConsensusState
		consensusState2 clientexported.ConsensusState
		evidence        clientexported.Misbehaviour
		consensusParams *abci.ConsensusParams
		timestamp       time.Time
		expPass         bool
	}{
		{
			"valid misbehavior evidence",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height, bothValsHash),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height, bothValsHash),
			types.Evidence{
				Header1:  types.CreateTestHeader(chainID, height, height, suite.now, bothValSet, bothValSet, bothSigners),
				Header2:  types.CreateTestHeader(chainID, height, height, suite.now.Add(time.Minute), bothValSet, bothValSet, bothSigners),
				ChainID:  chainID,
				ClientID: chainID,
			},
			simapp.DefaultConsensusParams,
			suite.now,
			true,
		},
		{
			"valid misbehavior at height greater than last consensusState",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height-1, bothValsHash),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height-1, bothValsHash),
			types.Evidence{
				Header1:  types.CreateTestHeader(chainID, height, height-1, suite.now, bothValSet, bothValSet, bothSigners),
				Header2:  types.CreateTestHeader(chainID, height, height-1, suite.now.Add(time.Minute), bothValSet, bothValSet, bothSigners),
				ChainID:  chainID,
				ClientID: chainID,
			},
			simapp.DefaultConsensusParams,
			suite.now,
			true,
		},
		{
			"valid misbehavior evidence with different trusted heights",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height-1, bothValsHash),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height-3, suite.valsHash),
			types.Evidence{
				Header1:  types.CreateTestHeader(chainID, height, height-1, suite.now, bothValSet, bothValSet, bothSigners),
				Header2:  types.CreateTestHeader(chainID, height, height-3, suite.now.Add(time.Minute), bothValSet, suite.valSet, bothSigners),
				ChainID:  chainID,
				ClientID: chainID,
			},
			simapp.DefaultConsensusParams,
			suite.now,
			true,
		},
		{
			"consensus state's valset hash different from evidence should still pass",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height-1, suite.valsHash),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height-1, suite.valsHash),
			types.Evidence{
				Header1:  types.CreateTestHeader(chainID, height, height-1, suite.now, bothValSet, suite.valSet, bothSigners),
				Header2:  types.CreateTestHeader(chainID, height, height-1, suite.now.Add(time.Minute), bothValSet, suite.valSet, bothSigners),
				ChainID:  chainID,
				ClientID: chainID,
			},
			simapp.DefaultConsensusParams,
			suite.now,
			true,
		},
		{
			"invalid misbehavior evidence with trusted height different from trusted consensus state",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height-1, bothValsHash),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height-3, suite.valsHash),
			types.Evidence{
				Header1:  types.CreateTestHeader(chainID, height, height-1, suite.now, bothValSet, bothValSet, bothSigners),
				Header2:  types.CreateTestHeader(chainID, height, height, suite.now.Add(time.Minute), bothValSet, suite.valSet, bothSigners),
				ChainID:  chainID,
				ClientID: chainID,
			},
			simapp.DefaultConsensusParams,
			suite.now,
			false,
		},
		{
			"invalid misbehavior evidence with trusted validators different from trusted consensus state",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height-1, bothValsHash),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height-3, suite.valsHash),
			types.Evidence{
				Header1:  types.CreateTestHeader(chainID, height, height-1, suite.now, bothValSet, bothValSet, bothSigners),
				Header2:  types.CreateTestHeader(chainID, height, height-3, suite.now.Add(time.Minute), bothValSet, bothValSet, bothSigners),
				ChainID:  chainID,
				ClientID: chainID,
			},
			simapp.DefaultConsensusParams,
			suite.now,
			false,
		},
		{
			"invalid tendermint client state",
			nil,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height, bothValsHash),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height, bothValsHash),
			types.Evidence{
				Header1:  types.CreateTestHeader(chainID, height, height, suite.now, bothValSet, bothValSet, bothSigners),
				Header2:  types.CreateTestHeader(chainID, height, height, suite.now.Add(time.Minute), bothValSet, altValSet, bothSigners),
				ChainID:  chainID,
				ClientID: chainID,
			},
			simapp.DefaultConsensusParams,
			suite.now,
			false,
		},
		{
			"already frozen client state",
			types.ClientState{FrozenHeight: 1},
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height, bothValsHash),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height, bothValsHash),
			types.Evidence{
				Header1:  types.CreateTestHeader(chainID, height, height, suite.now, bothValSet, bothValSet, bothSigners),
				Header2:  types.CreateTestHeader(chainID, height, height, suite.now.Add(time.Minute), bothValSet, bothValSet, bothSigners),
				ChainID:  chainID,
				ClientID: chainID,
			},
			simapp.DefaultConsensusParams,
			suite.now,
			false,
		},
		{
			"invalid tendermint consensus state",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			nil,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height, bothValsHash),
			types.Evidence{
				Header1:  types.CreateTestHeader(chainID, height, height, suite.now, bothValSet, bothValSet, bothSigners),
				Header2:  types.CreateTestHeader(chainID, height, height, suite.now.Add(time.Minute), bothValSet, bothValSet, bothSigners),
				ChainID:  chainID,
				ClientID: chainID,
			},
			simapp.DefaultConsensusParams,
			suite.now,
			false,
		},
		{
			"invalid tendermint misbehaviour evidence",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height, bothValsHash),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height, bothValsHash),
			nil,
			simapp.DefaultConsensusParams,
			suite.now,
			false,
		},
		{
			"rejected misbehaviour due to expired age",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height, bothValsHash),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height, bothValsHash),
			types.Evidence{
				Header1: types.CreateTestHeader(chainID, int64(2*height+uint64(simapp.DefaultConsensusParams.Evidence.MaxAgeNumBlocks)), height,
					suite.now, bothValSet, bothValSet, bothSigners),
				Header2: types.CreateTestHeader(chainID, int64(2*height+uint64(simapp.DefaultConsensusParams.Evidence.MaxAgeNumBlocks)), height,
					suite.now.Add(time.Minute), bothValSet, bothValSet, bothSigners),
				ChainID:  chainID,
				ClientID: chainID,
			},
			simapp.DefaultConsensusParams,
			suite.now.Add(2 * time.Minute).Add(simapp.DefaultConsensusParams.Evidence.MaxAgeDuration),
			false,
		},
		{
			"provided height > header height",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height, bothValsHash),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height, bothValsHash),
			types.Evidence{
				Header1:  types.CreateTestHeader(chainID, height, height-1, suite.now, bothValSet, bothValSet, bothSigners),
				Header2:  types.CreateTestHeader(chainID, height, height-1, suite.now.Add(time.Minute), bothValSet, bothValSet, bothSigners),
				ChainID:  chainID,
				ClientID: chainID,
			},
			simapp.DefaultConsensusParams,
			suite.now,
			false,
		},
		{
			"unbonding period expired",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			types.ConsensusState{Timestamp: time.Time{}, Root: commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), NextValidatorsHash: bothValsHash},
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height, bothValsHash),
			types.Evidence{
				Header1:  types.CreateTestHeader(chainID, height, height, suite.now, bothValSet, bothValSet, bothSigners),
				Header2:  types.CreateTestHeader(chainID, height, height, suite.now.Add(time.Minute), bothValSet, bothValSet, bothSigners),
				ChainID:  chainID,
				ClientID: chainID,
			},
			simapp.DefaultConsensusParams,
			suite.now,
			false,
		},
		{
			"trusted validators is incorrect for given consensus state",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height, bothValsHash),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height, bothValsHash),
			types.Evidence{
				Header1:  types.CreateTestHeader(chainID, height, height, suite.now, bothValSet, suite.valSet, bothSigners),
				Header2:  types.CreateTestHeader(chainID, height, height, suite.now.Add(time.Minute), bothValSet, suite.valSet, bothSigners),
				ChainID:  chainID,
				ClientID: chainID,
			},
			simapp.DefaultConsensusParams,
			suite.now,
			false,
		},
		{
			"first valset has too much change",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height, bothValsHash),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height, bothValsHash),
			types.Evidence{
				Header1:  types.CreateTestHeader(chainID, height, height, suite.now, altValSet, bothValSet, altSigners),
				Header2:  types.CreateTestHeader(chainID, height, height, suite.now.Add(time.Minute), bothValSet, bothValSet, bothSigners),
				ChainID:  chainID,
				ClientID: chainID,
			},
			simapp.DefaultConsensusParams,
			suite.now,
			false,
		},
		{
			"second valset has too much change",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height, bothValsHash),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height, bothValsHash),
			types.Evidence{
				Header1:  types.CreateTestHeader(chainID, height, height, suite.now, bothValSet, bothValSet, bothSigners),
				Header2:  types.CreateTestHeader(chainID, height, height, suite.now.Add(time.Minute), altValSet, bothValSet, altSigners),
				ChainID:  chainID,
				ClientID: chainID,
			},
			simapp.DefaultConsensusParams,
			suite.now,
			false,
		},
		{
			"both valsets have too much change",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height, bothValsHash),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height, bothValsHash),
			types.Evidence{
				Header1:  types.CreateTestHeader(chainID, height, height, suite.now, altValSet, bothValSet, altSigners),
				Header2:  types.CreateTestHeader(chainID, height, height, suite.now.Add(time.Minute), altValSet, bothValSet, altSigners),
				ChainID:  chainID,
				ClientID: chainID,
			},
			simapp.DefaultConsensusParams,
			suite.now,
			false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		clientState, err := tendermint.CheckMisbehaviourAndUpdateState(tc.clientState, tc.consensusState1, tc.consensusState2, tc.evidence, tc.timestamp, tc.consensusParams)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
			suite.Require().NotNil(clientState, "valid test case %d failed: %s", i, tc.name)
			suite.Require().True(clientState.IsFrozen(), "valid test case %d failed: %s", i, tc.name)
			suite.Require().Equal(uint64(tc.evidence.GetHeight()), clientState.GetFrozenHeight(),
				"valid test case %d failed: %s. Expected FrozenHeight %d got %d", tc.evidence.GetHeight(), clientState.GetFrozenHeight())
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
			suite.Require().Nil(clientState, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}
