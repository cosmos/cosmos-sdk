package types_test

import (
	"bytes"
	fmt "fmt"
	"time"

	"github.com/tendermint/tendermint/crypto/tmhash"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
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
			suite.now,
			false,
		},
		{
			"trusted consensus state does not exist",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			nil, // consensus state for trusted height - 1 does not exist in store
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height, bothValsHash),
			types.Evidence{
				Header1:  types.CreateTestHeader(chainID, height, height-1, suite.now, bothValSet, bothValSet, bothSigners),
				Header2:  types.CreateTestHeader(chainID, height, height, suite.now.Add(time.Minute), bothValSet, bothValSet, bothSigners),
				ChainID:  chainID,
				ClientID: chainID,
			},
			suite.now,
			false,
		},
		{
			"invalid tendermint misbehaviour evidence",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height, bothValsHash),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height, bothValsHash),
			nil,
			suite.now,
			false,
		},
		{
			"rejected misbehaviour due to expired age duration",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height, bothValsHash),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height, bothValsHash),
			types.Evidence{
				Header1:  types.CreateTestHeader(chainID, height, height, suite.now, bothValSet, bothValSet, bothSigners),
				Header2:  types.CreateTestHeader(chainID, height, height, suite.now.Add(time.Minute), bothValSet, bothValSet, bothSigners),
				ChainID:  chainID,
				ClientID: chainID,
			},
			suite.now.Add(2 * time.Minute).Add(simapp.DefaultConsensusParams.Evidence.MaxAgeDuration),
			false,
		},
		{
			"rejected misbehaviour due to expired block duration",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, uint64(height+simapp.DefaultConsensusParams.Evidence.MaxAgeNumBlocks+1), commitmenttypes.GetSDKSpecs()),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height, bothValsHash),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height, bothValsHash),
			types.Evidence{
				Header1:  types.CreateTestHeader(chainID, height, height, suite.now, bothValSet, bothValSet, bothSigners),
				Header2:  types.CreateTestHeader(chainID, height, height, suite.now.Add(time.Minute), bothValSet, bothValSet, bothSigners),
				ChainID:  chainID,
				ClientID: chainID,
			},
			suite.now.Add(time.Hour),
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
			suite.now,
			false,
		},
		{
			"unbonding period expired",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs()),
			&types.ConsensusState{Timestamp: time.Time{}, Root: commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), NextValidatorsHash: bothValsHash},
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), height, bothValsHash),
			types.Evidence{
				Header1:  types.CreateTestHeader(chainID, height, height, suite.now, bothValSet, bothValSet, bothSigners),
				Header2:  types.CreateTestHeader(chainID, height, height, suite.now.Add(time.Minute), bothValSet, bothValSet, bothSigners),
				ChainID:  chainID,
				ClientID: chainID,
			},
			suite.now.Add(ubdPeriod),
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
			suite.now,
			false,
		},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case: %s", tc.name), func() {

			if i != 10 {
				return
			}

			// reset suite to create fresh application state
			suite.SetupTest()

			// Set current timestamp in context
			ctx := suite.chainA.GetContext().WithBlockTime(tc.timestamp)
			ctx = ctx.WithConsensusParams(simapp.DefaultConsensusParams)

			// Set trusted consensus states in client store

			if tc.consensusState1 != nil {
				suite.chainA.App.IBCKeeper.ClientKeeper.SetClientConsensusState(ctx, clientID, tc.consensusState1.GetHeight(), tc.consensusState1)
			}
			if tc.consensusState2 != nil {
				suite.chainA.App.IBCKeeper.ClientKeeper.SetClientConsensusState(ctx, clientID, tc.consensusState2.GetHeight(), tc.consensusState2)
			}

			clientState, err := tc.clientState.CheckMisbehaviourAndUpdateState(
				ctx,
				suite.cdc,
				suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(ctx, clientID), // pass in clientID prefixed clientStore
				tc.evidence,
			)

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
		})
	}
}
