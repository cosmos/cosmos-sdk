package keeper_test

import (
	"fmt"
	"time"

	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	localhosttypes "github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
	ibctestingmock "github.com/cosmos/cosmos-sdk/x/ibc/testing/mock"
)

const (
	invalidClientType exported.ClientType = 0
)

func (suite *KeeperTestSuite) TestCreateClient() {
	suite.chainA.App.IBCKeeper.ClientKeeper.SetClientType(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], exported.Tendermint)

	cases := []struct {
		msg      string
		clientID string
		expPass  bool
		expPanic bool
	}{
		{"success", suite.chainA.ClientIDs[0], true, false},
		{"client ID exists", suite.chainA.ClientIDs[0], false, false},
		{"client type exists", suite.chainA.ClientIDs[0], false, true},
	}

	for i, tc := range cases {
		tc := tc
		i := i
		consensusState := suite.chainA.ConsensusStateFromCurrentHeader()
		if tc.expPanic {
			suite.Require().Panics(func() {
				clientState := ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, ibctesting.ClientHeight, commitmenttypes.GetSDKSpecs(), false, false)
				suite.chainA.App.IBCKeeper.ClientKeeper.CreateClient(suite.chainA.GetContext(), tc.clientID, clientState, consensusState)
			}, "Msg %d didn't panic: %s", i, tc.msg)
		} else {
			clientState := ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, ibctesting.ClientHeight, commitmenttypes.GetSDKSpecs(), false, false)
			if tc.expPass {
				suite.Require().NotNil(clientState, "valid test case %d failed: %s", i, tc.msg)
			}
			// If we were able to NewClientState clientstate successfully, try persisting it to state
			_, err := suite.chainA.App.IBCKeeper.ClientKeeper.CreateClient(suite.chainA.GetContext(), tc.clientID, clientState, consensusState)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		}
	}
}

func (suite *KeeperTestSuite) TestUpdateClientTendermint() {
	// Must create header creation functions since suite.header gets recreated on each test case
	createFutureUpdateFn := func(s *KeeperTestSuite) *ibctmtypes.Header {
		header := suite.chainA.CreateTMClientHeader()
		header.Header.Height = int64(header.GetHeight().GetEpochHeight() + 3)
		header.Header.Time = header.GetTime().Add(time.Hour)
		return header
	}
	createPastUpdateFn := func(s *KeeperTestSuite) *ibctmtypes.Header {
		header := suite.chainA.CreateTMClientHeader()
		header.Header.Height = int64(header.GetHeight().GetEpochHeight() - 2)
		header.TrustedHeight.EpochHeight = header.GetHeight().GetEpochHeight() - 4
		return header
	}
	var (
		updateHeader *ibctmtypes.Header
		clientState  *ibctmtypes.ClientState
	)

	cases := []struct {
		name     string
		malleate func() error
		expPass  bool
	}{
		{"valid update", func() error {
			consensusState := suite.chainA.ConsensusStateFromCurrentHeader()
			clientState = ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, ibctesting.ClientHeight, commitmenttypes.GetSDKSpecs(), false, false)
			_, err := suite.chainA.App.IBCKeeper.ClientKeeper.CreateClient(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState, consensusState)

			// store intermediate consensus state to check that trustedHeight does not need to be highest consensus state before header height
			incrementedClientHeight := ibctesting.ClientHeight.Increment()
			intermediateConsState := &ibctmtypes.ConsensusState{
				Height:             incrementedClientHeight,
				Timestamp:          suite.chainA.CurrentHeader.Time.Add(time.Minute),
				NextValidatorsHash: suite.chainA.CurrentHeader.NextValidatorsHash,
			}
			suite.chainA.App.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], incrementedClientHeight, intermediateConsState)

			clientState.LatestHeight = incrementedClientHeight
			suite.chainA.App.IBCKeeper.ClientKeeper.SetClientState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState)

			updateHeader = createFutureUpdateFn(suite)
			return err
		}, true},
		{"valid past update", func() error {
			consensusState := suite.chainA.ConsensusStateFromCurrentHeader()
			clientState = ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, ibctesting.ClientHeight, commitmenttypes.GetSDKSpecs(), false, false)
			_, err := suite.chainA.App.IBCKeeper.ClientKeeper.CreateClient(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState, consensusState)
			suite.Require().NoError(err)

			height1 := types.NewHeight(0, 1)

			// store previous consensus state
			prevConsState := &ibctmtypes.ConsensusState{
				Height:             height1,
				Timestamp:          suite.chainA.CurrentHeader.Time.Add(-time.Hour),
				NextValidatorsHash: suite.chainA.CurrentHeader.NextValidatorsHash,
			}
			suite.chainA.App.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], height1, prevConsState)

			height2 := types.NewHeight(0, 2)

			// store intermediate consensus state to check that trustedHeight does not need to be hightest consensus state before header height
			intermediateConsState := &ibctmtypes.ConsensusState{
				Height:             height2,
				Timestamp:          suite.chainA.CurrentHeader.Time.Add(-time.Hour).Add(time.Minute),
				NextValidatorsHash: suite.chainA.CurrentHeader.NextValidatorsHash,
			}
			suite.chainA.App.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], height2, intermediateConsState)

			// updateHeader will fill in consensus state between prevConsState and suite.consState
			// clientState should not be updated
			updateHeader = createPastUpdateFn(suite)
			return nil
		}, true},
		{"client type not found", func() error {
			updateHeader = createFutureUpdateFn(suite)

			return nil
		}, false},
		{"client type and header type mismatch", func() error {
			suite.chainA.App.IBCKeeper.ClientKeeper.SetClientType(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], invalidClientType)
			updateHeader = createFutureUpdateFn(suite)

			return nil
		}, false},
		{"client state not found", func() error {
			suite.chainA.App.IBCKeeper.ClientKeeper.SetClientType(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], exported.Tendermint)
			updateHeader = createFutureUpdateFn(suite)

			return nil
		}, false},
		{"consensus state not found", func() error {
			clientState = ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, ibctesting.ClientHeight, commitmenttypes.GetSDKSpecs(), false, false)
			suite.chainA.App.IBCKeeper.ClientKeeper.SetClientState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState)
			suite.chainA.App.IBCKeeper.ClientKeeper.SetClientType(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], exported.Tendermint)
			updateHeader = createFutureUpdateFn(suite)

			return nil
		}, false},
		{"frozen client before update", func() error {
			clientState = &ibctmtypes.ClientState{FrozenHeight: types.NewHeight(0, 1), LatestHeight: ibctesting.ClientHeight}
			suite.chainA.App.IBCKeeper.ClientKeeper.SetClientState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState)
			suite.chainA.App.IBCKeeper.ClientKeeper.SetClientType(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], exported.Tendermint)
			updateHeader = createFutureUpdateFn(suite)

			return nil
		}, false},
		{"valid past update before client was frozen", func() error {
			consensusState := suite.chainA.ConsensusStateFromCurrentHeader()
			clientState = ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, ibctesting.ClientHeight, commitmenttypes.GetSDKSpecs(), false, false)
			clientState.FrozenHeight = types.NewHeight(0, ibctesting.ClientHeight.EpochHeight-1)
			_, err := suite.chainA.App.IBCKeeper.ClientKeeper.CreateClient(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState, consensusState)
			suite.Require().NoError(err)

			height1 := types.NewHeight(0, 1)

			// store previous consensus state
			prevConsState := &ibctmtypes.ConsensusState{
				Height:             height1,
				Timestamp:          suite.chainA.CurrentHeader.Time.Add(-time.Hour),
				NextValidatorsHash: suite.chainA.CurrentHeader.NextValidatorsHash,
			}
			suite.chainA.App.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], height1, prevConsState)

			// updateHeader will fill in consensus state between prevConsState and suite.consState
			// clientState should not be updated
			updateHeader = createPastUpdateFn(suite)
			return nil
		}, true},
		{"invalid header", func() error {
			consensusState := suite.chainA.ConsensusStateFromCurrentHeader()
			clientState = ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, ibctesting.ClientHeight, commitmenttypes.GetSDKSpecs(), false, false)
			_, err := suite.chainA.App.IBCKeeper.ClientKeeper.CreateClient(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState, consensusState)
			suite.Require().NoError(err)
			updateHeader = createPastUpdateFn(suite)

			return nil
		}, false},
	}

	for i, tc := range cases {
		tc := tc
		i := i
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()

			err := tc.malleate()
			suite.Require().NoError(err)

			ctx := suite.chainA.GetContext().WithBlockTime(updateHeader.GetTime().Add(time.Minute))

			updatedClientState, err := suite.chainA.App.IBCKeeper.ClientKeeper.UpdateClient(ctx, suite.chainA.ClientIDs[0], updateHeader)

			if tc.expPass {
				suite.Require().NoError(err, err)

				expConsensusState := &ibctmtypes.ConsensusState{
					Height:             updateHeader.GetHeight().(types.Height),
					Timestamp:          updateHeader.GetTime(),
					Root:               commitmenttypes.NewMerkleRoot(updateHeader.Header.GetAppHash()),
					NextValidatorsHash: updateHeader.Header.NextValidatorsHash,
				}

				newClientState, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientState(ctx, suite.chainA.ClientIDs[0])
				suite.Require().True(found, "valid test case %d failed: %s", i, tc.name)

				consensusState, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientConsensusState(ctx, suite.chainA.ClientIDs[0], updateHeader.GetHeight())
				suite.Require().True(found, "valid test case %d failed: %s", i, tc.name)
				// check returned client state is same as client state in store
				suite.Require().Equal(updatedClientState, newClientState, "updatedClient state not persisted correctly")

				// Determine if clientState should be updated or not
				if updateHeader.GetHeight().GT(clientState.GetLatestHeight()) {
					// Header Height is greater than clientState latest Height, clientState should be updated with header.GetHeight()
					suite.Require().Equal(updateHeader.GetHeight(), updatedClientState.GetLatestHeight(), "clientstate height did not update")
				} else {
					// Update will add past consensus state, clientState should not be updated at all
					suite.Require().Equal(clientState.GetLatestHeight(), updatedClientState.GetLatestHeight(), "client state height updated for past header")
				}

				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
				suite.Require().Equal(expConsensusState, consensusState, "consensus state should have been updated on case %s", tc.name)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestUpdateClientLocalhost() {
	var localhostClient exported.ClientState = localhosttypes.NewClientState(suite.chainA.ChainID, types.NewHeight(0, uint64(suite.chainA.GetContext().BlockHeight())))

	ctx := suite.chainA.GetContext().WithBlockHeight(suite.chainA.GetContext().BlockHeight() + 1)

	updatedClientState, err := suite.chainA.App.IBCKeeper.ClientKeeper.UpdateClient(ctx, exported.ClientTypeLocalHost, nil)
	suite.Require().NoError(err, err)
	suite.Require().Equal(localhostClient.GetLatestHeight().(types.Height).Increment(), updatedClientState.GetLatestHeight())
}

func (suite *KeeperTestSuite) TestCheckMisbehaviourAndUpdateState() {
	altPrivVal := ibctestingmock.NewPV()
	altPubKey, err := altPrivVal.GetPubKey()
	suite.Require().NoError(err)
	altVal := tmtypes.NewValidator(altPubKey, 4)

	// Create bothValSet with both suite validator and altVal
	bothValSet := tmtypes.NewValidatorSet(append(suite.chainA.Vals.Validators, altVal))
	bothValsHash := bothValSet.Hash()
	// Create alternative validator set with only altVal
	altValSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{altVal})

	// Create signer array and ensure it is in same order as bothValSet
	_, suiteVal := suite.chainA.Vals.GetByIndex(0)
	bothSigners := ibctmtypes.CreateSortedSignerArray(altPrivVal, suite.chainA.Signers[0], altVal, suiteVal)

	altSigners := []tmtypes.PrivValidator{altPrivVal}

	// Create valid Misbehaviour by making a duplicate header that signs over different block time
	altTime := suite.chainA.GetContext().BlockTime().Add(time.Minute)

	heightPlus3 := types.NewHeight(0, ibctesting.Height+3)

	consensusState, ok := suite.chainA.ConsensusStateFromCurrentHeader().(ibctmtypes.ConsensusState)
	suite.Require().True(ok)

	testCases := []struct {
		name         string
		misbehaviour *ibctmtypes.Misbehaviour
		malleate     func() error
		expPass      bool
	}{
		{
			"trusting period misbehavior should pass",
			&ibctmtypes.Misbehaviour{
				Header1:  ibctmtypes.CreateTestHeader(suite.chainA.ChainID, ibctesting.Height, ibctesting.Height, altTime, bothValSet, bothValSet, bothSigners),
				Header2:  ibctmtypes.CreateTestHeader(suite.chainA.ChainID, ibctesting.Height, ibctesting.Height, suite.chainA.GetContext().BlockTime(), bothValSet, bothValSet, bothSigners),
				ChainId:  suite.chainA.ChainID,
				ClientId: suite.chainA.ClientIDs[0],
			},
			func() error {
				consensusState.NextValidatorsHash = bothValsHash
				clientState := ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, ibctesting.ClientHeight, commitmenttypes.GetSDKSpecs(), false, false)
				_, err := suite.chainA.App.IBCKeeper.ClientKeeper.CreateClient(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState, consensusState)

				return err
			},
			true,
		},
		{
			"misbehavior at later height should pass",
			&ibctmtypes.Misbehaviour{
				Header1:  ibctmtypes.CreateTestHeader(suite.chainA.ChainID, ibctesting.Height+5, ibctesting.Height, altTime, bothValSet, suite.chainA.Vals, bothSigners),
				Header2:  ibctmtypes.CreateTestHeader(suite.chainA.ChainID, ibctesting.Height+5, ibctesting.Height, suite.chainA.GetContext().BlockTime(), bothValSet, suite.chainA.Vals, bothSigners),
				ChainId:  suite.chainA.ChainID,
				ClientId: suite.chainA.ClientIDs[0],
			},
			func() error {
				consensusState.NextValidatorsHash = suite.chainA.Vals.Hash()
				clientState := ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, ibctesting.ClientHeight, commitmenttypes.GetSDKSpecs(), false, false)
				_, err := suite.chainA.App.IBCKeeper.ClientKeeper.CreateClient(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState, consensusState)

				// store intermediate consensus state to check that trustedHeight does not need to be highest consensus state before header height
				intermediateConsState := &ibctmtypes.ConsensusState{
					Height:             heightPlus3,
					Timestamp:          suite.chainA.CurrentHeader.Time.Add(time.Minute),
					NextValidatorsHash: suite.chainA.CurrentHeader.NextValidatorsHash,
				}
				suite.chainA.App.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], heightPlus3, intermediateConsState)

				clientState.LatestHeight = heightPlus3
				suite.chainA.App.IBCKeeper.ClientKeeper.SetClientState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState)

				return err
			},
			true,
		},
		{
			"misbehavior at later height with different trusted heights should pass",
			&ibctmtypes.Misbehaviour{
				Header1:  ibctmtypes.CreateTestHeader(suite.chainA.ChainID, ibctesting.Height+5, ibctesting.Height, altTime, bothValSet, suite.chainA.Vals, bothSigners),
				Header2:  ibctmtypes.CreateTestHeader(suite.chainA.ChainID, ibctesting.Height+5, ibctesting.Height+3, suite.chainA.GetContext().BlockTime(), bothValSet, bothValSet, bothSigners),
				ChainId:  suite.chainA.ChainID,
				ClientId: suite.chainA.ClientIDs[0],
			},
			func() error {
				consensusState.NextValidatorsHash = suite.chainA.Vals.Hash()
				clientState := ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, ibctesting.ClientHeight, commitmenttypes.GetSDKSpecs(), false, false)
				_, err := suite.chainA.App.IBCKeeper.ClientKeeper.CreateClient(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState, consensusState)

				// store trusted consensus state for Header2
				intermediateConsState := &ibctmtypes.ConsensusState{
					Height:             heightPlus3,
					Timestamp:          suite.chainA.CurrentHeader.Time.Add(time.Minute),
					NextValidatorsHash: bothValsHash,
				}
				suite.chainA.App.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], heightPlus3, intermediateConsState)

				clientState.LatestHeight = heightPlus3
				suite.chainA.App.IBCKeeper.ClientKeeper.SetClientState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState)

				return err
			},
			true,
		},
		{
			"misbehaviour fails validatebasic",
			&ibctmtypes.Misbehaviour{
				Header1:  ibctmtypes.CreateTestHeader(suite.chainA.ChainID, ibctesting.Height+1, ibctesting.Height, altTime, bothValSet, bothValSet, bothSigners),
				Header2:  ibctmtypes.CreateTestHeader(suite.chainA.ChainID, ibctesting.Height, ibctesting.Height, suite.chainA.GetContext().BlockTime(), bothValSet, bothValSet, bothSigners),
				ChainId:  suite.chainA.ChainID,
				ClientId: suite.chainA.ClientIDs[0],
			},
			func() error {
				consensusState.NextValidatorsHash = bothValsHash
				clientState := ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, ibctesting.ClientHeight, commitmenttypes.GetSDKSpecs(), false, false)
				_, err := suite.chainA.App.IBCKeeper.ClientKeeper.CreateClient(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState, consensusState)

				return err
			},
			false,
		},
		{
			"trusted ConsensusState1 not found",
			&ibctmtypes.Misbehaviour{
				Header1:  ibctmtypes.CreateTestHeader(suite.chainA.ChainID, ibctesting.Height+5, ibctesting.Height+3, altTime, bothValSet, bothValSet, bothSigners),
				Header2:  ibctmtypes.CreateTestHeader(suite.chainA.ChainID, ibctesting.Height+5, ibctesting.Height, suite.chainA.GetContext().BlockTime(), bothValSet, suite.chainA.Vals, bothSigners),
				ChainId:  suite.chainA.ChainID,
				ClientId: suite.chainA.ClientIDs[0],
			},
			func() error {
				consensusState.NextValidatorsHash = suite.chainA.Vals.Hash()
				clientState := ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, ibctesting.ClientHeight, commitmenttypes.GetSDKSpecs(), false, false)
				_, err := suite.chainA.App.IBCKeeper.ClientKeeper.CreateClient(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState, consensusState)
				// intermediate consensus state at height + 3 is not created
				return err
			},
			false,
		},
		{
			"trusted ConsensusState2 not found",
			&ibctmtypes.Misbehaviour{
				Header1:  ibctmtypes.CreateTestHeader(suite.chainA.ChainID, ibctesting.Height+5, ibctesting.Height, altTime, bothValSet, suite.chainA.Vals, bothSigners),
				Header2:  ibctmtypes.CreateTestHeader(suite.chainA.ChainID, ibctesting.Height+5, ibctesting.Height+3, suite.chainA.GetContext().BlockTime(), bothValSet, bothValSet, bothSigners),
				ChainId:  suite.chainA.ChainID,
				ClientId: suite.chainA.ClientIDs[0],
			},
			func() error {

				consensusState.NextValidatorsHash = suite.chainA.Vals.Hash()
				clientState := ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, ibctesting.ClientHeight, commitmenttypes.GetSDKSpecs(), false, false)
				_, err := suite.chainA.App.IBCKeeper.ClientKeeper.CreateClient(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState, consensusState)
				// intermediate consensus state at height + 3 is not created
				return err
			},
			false,
		},
		{
			"client state not found",
			&ibctmtypes.Misbehaviour{},
			func() error { return nil },
			false,
		},
		{
			"client already frozen at earlier height",
			&ibctmtypes.Misbehaviour{
				Header1:  ibctmtypes.CreateTestHeader(suite.chainA.ChainID, ibctesting.Height, ibctesting.Height, altTime, bothValSet, bothValSet, bothSigners),
				Header2:  ibctmtypes.CreateTestHeader(suite.chainA.ChainID, ibctesting.Height, ibctesting.Height, suite.chainA.GetContext().BlockTime(), bothValSet, bothValSet, bothSigners),
				ChainId:  suite.chainA.ChainID,
				ClientId: suite.chainA.ClientIDs[0],
			},
			func() error {
				consensusState.NextValidatorsHash = bothValsHash
				clientState := ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, ibctesting.ClientHeight, commitmenttypes.GetSDKSpecs(), false, false)
				_, err := suite.chainA.App.IBCKeeper.ClientKeeper.CreateClient(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState, consensusState)

				clientState.FrozenHeight = types.NewHeight(0, 1)
				suite.chainA.App.IBCKeeper.ClientKeeper.SetClientState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState)

				return err
			},
			false,
		},
		{
			"misbehaviour check failed",
			&ibctmtypes.Misbehaviour{
				Header1:  ibctmtypes.CreateTestHeader(suite.chainA.ChainID, ibctesting.Height, ibctesting.Height, altTime, bothValSet, bothValSet, bothSigners),
				Header2:  ibctmtypes.CreateTestHeader(suite.chainA.ChainID, ibctesting.Height, ibctesting.Height, suite.chainA.GetContext().BlockTime(), altValSet, bothValSet, altSigners),
				ChainId:  suite.chainA.ChainID,
				ClientId: suite.chainA.ClientIDs[0],
			},
			func() error {
				clientState := ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, ibctesting.ClientHeight, commitmenttypes.GetSDKSpecs(), false, false)
				if err != nil {
					return err
				}
				_, err = suite.chainA.App.IBCKeeper.ClientKeeper.CreateClient(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState, consensusState)

				return err
			},
			false,
		},
	}

	for i, tc := range testCases {
		tc := tc
		i := i
		suite.Run(tc.name, func() {
			suite.SetupTest() // reset

			err := tc.malleate()
			suite.Require().NoError(err)

			err = suite.chainA.App.IBCKeeper.ClientKeeper.CheckMisbehaviourAndUpdateState(suite.chainA.GetContext(), tc.misbehaviour)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)

				clientState, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0])
				suite.Require().True(found, "valid test case %d failed: %s", i, tc.name)
				suite.Require().True(clientState.IsFrozen(), "valid test case %d failed: %s", i, tc.name)
				suite.Require().Equal(tc.misbehaviour.GetHeight(), clientState.GetFrozenHeight(),
					"valid test case %d failed: %s. Expected FrozenHeight %s got %s", tc.misbehaviour.GetHeight(), clientState.GetFrozenHeight())
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
			}
		})
	}
}
