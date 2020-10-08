package types_test

import (
	"fmt"
	"time"

	"github.com/tendermint/tendermint/crypto/tmhash"
	tmtypes "github.com/tendermint/tendermint/types"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/23-commitment/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
	ibctestingmock "github.com/cosmos/cosmos-sdk/x/ibc/testing/mock"
)

func (suite *TendermintTestSuite) TestCheckMisbehaviourAndUpdateState() {
	altPrivVal := ibctestingmock.NewPV()
	altPubKey, err := altPrivVal.GetPubKey()
	suite.Require().NoError(err)

	altVal := tmtypes.NewValidator(altPubKey.(cryptotypes.IntoTmPubKey).AsTmPubKey(), 4)

	// Create bothValSet with both suite validator and altVal
	bothValSet := tmtypes.NewValidatorSet(append(suite.valSet.Validators, altVal))
	bothValsHash := bothValSet.Hash()
	// Create alternative validator set with only altVal
	altValSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{altVal})

	_, suiteVal := suite.valSet.GetByIndex(0)

	// Create signer array and ensure it is in same order as bothValSet
	bothSigners := types.CreateSortedSignerArray(altPrivVal, suite.privVal, altVal, suiteVal)

	altSigners := []tmtypes.PrivValidator{altPrivVal}

	versionHeight := int64(height.VersionHeight)
	heightMinus1 := clienttypes.NewHeight(height.VersionNumber, height.VersionHeight-1)
	heightMinus3 := clienttypes.NewHeight(height.VersionNumber, height.VersionHeight-3)

	testCases := []struct {
		name            string
		clientState     exported.ClientState
		consensusState1 exported.ConsensusState
		height1         clienttypes.Height
		consensusState2 exported.ConsensusState
		height2         clienttypes.Height
		misbehaviour    exported.Misbehaviour
		timestamp       time.Time
		expPass         bool
	}{
		{
			"valid misbehavior misbehaviour",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			&types.Misbehaviour{
				Header1:  types.CreateTestHeader(chainID, height, height, suite.now, bothValSet, bothValSet, bothSigners),
				Header2:  types.CreateTestHeader(chainID, height, height, suite.now.Add(time.Minute), bothValSet, bothValSet, bothSigners),
				ChainId:  chainID,
				ClientId: chainID,
			},
			suite.now,
			true,
		},
		{
			"valid misbehavior at height greater than last consensusState",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			heightMinus1,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			heightMinus1,
			&types.Misbehaviour{
				Header1:  types.CreateTestHeader(chainID, height, heightMinus1, suite.now, bothValSet, bothValSet, bothSigners),
				Header2:  types.CreateTestHeader(chainID, height, heightMinus1, suite.now.Add(time.Minute), bothValSet, bothValSet, bothSigners),
				ChainId:  chainID,
				ClientId: chainID,
			},
			suite.now,
			true,
		},
		{
			"valid misbehaviour with different trusted heights",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			heightMinus1,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), suite.valsHash),
			heightMinus3,
			&types.Misbehaviour{
				Header1:  types.CreateTestHeader(chainID, height, heightMinus1, suite.now, bothValSet, bothValSet, bothSigners),
				Header2:  types.CreateTestHeader(chainID, height, heightMinus3, suite.now.Add(time.Minute), bothValSet, suite.valSet, bothSigners),
				ChainId:  chainID,
				ClientId: chainID,
			},
			suite.now,
			true,
		},
		{
			"valid misbehaviour at a previous version",
			types.NewClientState(chainIDVersion1, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, clienttypes.NewHeight(1, 1), ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			heightMinus1,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), suite.valsHash),
			heightMinus3,
			&types.Misbehaviour{
				Header1:  types.CreateTestHeader(chainIDVersion0, height, heightMinus1, suite.now, bothValSet, bothValSet, bothSigners),
				Header2:  types.CreateTestHeader(chainIDVersion0, height, heightMinus3, suite.now.Add(time.Minute), bothValSet, suite.valSet, bothSigners),
				ChainId:  chainIDVersion0,
				ClientId: chainID,
			},
			suite.now,
			true,
		},
		{
			"valid misbehaviour at a future version",
			types.NewClientState(chainIDVersion0, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			heightMinus1,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), suite.valsHash),
			heightMinus3,
			&types.Misbehaviour{
				Header1:  types.CreateTestHeader(chainIDVersion0, clienttypes.NewHeight(1, 3), heightMinus1, suite.now, bothValSet, bothValSet, bothSigners),
				Header2:  types.CreateTestHeader(chainIDVersion0, clienttypes.NewHeight(1, 3), heightMinus3, suite.now.Add(time.Minute), bothValSet, suite.valSet, bothSigners),
				ChainId:  chainIDVersion0,
				ClientId: chainID,
			},
			suite.now,
			true,
		},
		{
			"valid misbehaviour with trusted heights at a previous version",
			types.NewClientState(chainIDVersion1, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, clienttypes.NewHeight(1, 1), ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			heightMinus1,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), suite.valsHash),
			heightMinus3,
			&types.Misbehaviour{
				Header1:  types.CreateTestHeader(chainIDVersion1, clienttypes.NewHeight(1, 1), heightMinus1, suite.now, bothValSet, bothValSet, bothSigners),
				Header2:  types.CreateTestHeader(chainIDVersion1, clienttypes.NewHeight(1, 1), heightMinus3, suite.now.Add(time.Minute), bothValSet, suite.valSet, bothSigners),
				ChainId:  chainIDVersion1,
				ClientId: chainID,
			},
			suite.now,
			true,
		},
		{
			"consensus state's valset hash different from misbehaviour should still pass",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), suite.valsHash),
			height,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), suite.valsHash),
			height,
			&types.Misbehaviour{
				Header1:  types.CreateTestHeader(chainID, height, height, suite.now, bothValSet, suite.valSet, bothSigners),
				Header2:  types.CreateTestHeader(chainID, height, height, suite.now.Add(time.Minute), bothValSet, suite.valSet, bothSigners),
				ChainId:  chainID,
				ClientId: chainID,
			},
			suite.now,
			true,
		},
		{
			"invalid misbehavior misbehaviour from different chain",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			&types.Misbehaviour{
				Header1:  types.CreateTestHeader("ethermint", height, height, suite.now, bothValSet, bothValSet, bothSigners),
				Header2:  types.CreateTestHeader("ethermint", height, height, suite.now.Add(time.Minute), bothValSet, bothValSet, bothSigners),
				ChainId:  "ethermint",
				ClientId: chainID,
			},
			suite.now,
			false,
		},
		{
			"invalid misbehavior misbehaviour with trusted height different from trusted consensus state",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			heightMinus1,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), suite.valsHash),
			heightMinus3,
			&types.Misbehaviour{
				Header1:  types.CreateTestHeader(chainID, height, heightMinus1, suite.now, bothValSet, bothValSet, bothSigners),
				Header2:  types.CreateTestHeader(chainID, height, height, suite.now.Add(time.Minute), bothValSet, suite.valSet, bothSigners),
				ChainId:  chainID,
				ClientId: chainID,
			},
			suite.now,
			false,
		},
		{
			"invalid misbehavior misbehaviour with trusted validators different from trusted consensus state",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			heightMinus1,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), suite.valsHash),
			heightMinus3,
			&types.Misbehaviour{
				Header1:  types.CreateTestHeader(chainID, height, heightMinus1, suite.now, bothValSet, bothValSet, bothSigners),
				Header2:  types.CreateTestHeader(chainID, height, heightMinus3, suite.now.Add(time.Minute), bothValSet, bothValSet, bothSigners),
				ChainId:  chainID,
				ClientId: chainID,
			},
			suite.now,
			false,
		},
		{
			"already frozen client state",
			&types.ClientState{FrozenHeight: clienttypes.NewHeight(0, 1)},
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			&types.Misbehaviour{
				Header1:  types.CreateTestHeader(chainID, height, height, suite.now, bothValSet, bothValSet, bothSigners),
				Header2:  types.CreateTestHeader(chainID, height, height, suite.now.Add(time.Minute), bothValSet, bothValSet, bothSigners),
				ChainId:  chainID,
				ClientId: chainID,
			},
			suite.now,
			false,
		},
		{
			"trusted consensus state does not exist",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			nil, // consensus state for trusted height - 1 does not exist in store
			clienttypes.Height{},
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			&types.Misbehaviour{
				Header1:  types.CreateTestHeader(chainID, height, heightMinus1, suite.now, bothValSet, bothValSet, bothSigners),
				Header2:  types.CreateTestHeader(chainID, height, height, suite.now.Add(time.Minute), bothValSet, bothValSet, bothSigners),
				ChainId:  chainID,
				ClientId: chainID,
			},
			suite.now,
			false,
		},
		{
			"invalid tendermint misbehaviour",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			nil,
			suite.now,
			false,
		},
		{
			"rejected misbehaviour due to expired age duration",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			&types.Misbehaviour{
				Header1:  types.CreateTestHeader(chainID, height, height, suite.now, bothValSet, bothValSet, bothSigners),
				Header2:  types.CreateTestHeader(chainID, height, height, suite.now.Add(time.Minute), bothValSet, bothValSet, bothSigners),
				ChainId:  chainID,
				ClientId: chainID,
			},
			suite.now.Add(2 * time.Minute).Add(simapp.DefaultConsensusParams.Evidence.MaxAgeDuration),
			false,
		},
		{
			"rejected misbehaviour due to expired block duration",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, clienttypes.NewHeight(0, uint64(versionHeight+simapp.DefaultConsensusParams.Evidence.MaxAgeNumBlocks+1)), ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			&types.Misbehaviour{
				Header1:  types.CreateTestHeader(chainID, height, height, suite.now, bothValSet, bothValSet, bothSigners),
				Header2:  types.CreateTestHeader(chainID, height, height, suite.now.Add(time.Minute), bothValSet, bothValSet, bothSigners),
				ChainId:  chainID,
				ClientId: chainID,
			},
			suite.now.Add(time.Hour),
			false,
		},
		{
			"provided height > header height",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			&types.Misbehaviour{
				Header1:  types.CreateTestHeader(chainID, height, heightMinus1, suite.now, bothValSet, bothValSet, bothSigners),
				Header2:  types.CreateTestHeader(chainID, height, heightMinus1, suite.now.Add(time.Minute), bothValSet, bothValSet, bothSigners),
				ChainId:  chainID,
				ClientId: chainID,
			},
			suite.now,
			false,
		},
		{
			"unbonding period expired",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(time.Time{}, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			heightMinus1,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			&types.Misbehaviour{
				Header1:  types.CreateTestHeader(chainID, height, heightMinus1, suite.now, bothValSet, bothValSet, bothSigners),
				Header2:  types.CreateTestHeader(chainID, height, height, suite.now.Add(time.Minute), bothValSet, bothValSet, bothSigners),
				ChainId:  chainID,
				ClientId: chainID,
			},
			suite.now.Add(ubdPeriod),
			false,
		},
		{
			"trusted validators is incorrect for given consensus state",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			&types.Misbehaviour{
				Header1:  types.CreateTestHeader(chainID, height, height, suite.now, bothValSet, suite.valSet, bothSigners),
				Header2:  types.CreateTestHeader(chainID, height, height, suite.now.Add(time.Minute), bothValSet, suite.valSet, bothSigners),
				ChainId:  chainID,
				ClientId: chainID,
			},
			suite.now,
			false,
		},
		{
			"first valset has too much change",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			&types.Misbehaviour{
				Header1:  types.CreateTestHeader(chainID, height, height, suite.now, altValSet, bothValSet, altSigners),
				Header2:  types.CreateTestHeader(chainID, height, height, suite.now.Add(time.Minute), bothValSet, bothValSet, bothSigners),
				ChainId:  chainID,
				ClientId: chainID,
			},
			suite.now,
			false,
		},
		{
			"second valset has too much change",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			&types.Misbehaviour{
				Header1:  types.CreateTestHeader(chainID, height, height, suite.now, bothValSet, bothValSet, bothSigners),
				Header2:  types.CreateTestHeader(chainID, height, height, suite.now.Add(time.Minute), altValSet, bothValSet, altSigners),
				ChainId:  chainID,
				ClientId: chainID,
			},
			suite.now,
			false,
		},
		{
			"both valsets have too much change",
			types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), upgradePath, false, false),
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			types.NewConsensusState(suite.now, commitmenttypes.NewMerkleRoot(tmhash.Sum([]byte("app_hash"))), bothValsHash),
			height,
			&types.Misbehaviour{
				Header1:  types.CreateTestHeader(chainID, height, height, suite.now, altValSet, bothValSet, altSigners),
				Header2:  types.CreateTestHeader(chainID, height, height, suite.now.Add(time.Minute), altValSet, bothValSet, altSigners),
				ChainId:  chainID,
				ClientId: chainID,
			},
			suite.now,
			false,
		},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case: %s", tc.name), func() {
			// reset suite to create fresh application state
			suite.SetupTest()

			// Set current timestamp in context
			ctx := suite.chainA.GetContext().WithBlockTime(tc.timestamp)

			// Set trusted consensus states in client store

			if tc.consensusState1 != nil {
				suite.chainA.App.IBCKeeper.ClientKeeper.SetClientConsensusState(ctx, clientID, tc.height1, tc.consensusState1)
			}
			if tc.consensusState2 != nil {
				suite.chainA.App.IBCKeeper.ClientKeeper.SetClientConsensusState(ctx, clientID, tc.height2, tc.consensusState2)
			}

			clientState, err := tc.clientState.CheckMisbehaviourAndUpdateState(
				ctx,
				suite.cdc,
				suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(ctx, clientID), // pass in clientID prefixed clientStore
				tc.misbehaviour,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
				suite.Require().NotNil(clientState, "valid test case %d failed: %s", i, tc.name)
				suite.Require().True(clientState.IsFrozen(), "valid test case %d failed: %s", i, tc.name)
				suite.Require().Equal(tc.misbehaviour.GetHeight(), clientState.GetFrozenHeight(),
					"valid test case %d failed: %s. Expected FrozenHeight %s got %s", tc.misbehaviour.GetHeight(), clientState.GetFrozenHeight())
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
				suite.Require().Nil(clientState, "invalid test case %d passed: %s", i, tc.name)
			}
		})
	}
}
