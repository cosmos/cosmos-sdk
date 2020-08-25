package keeper_test

import (
	"fmt"
	"time"

	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	localhosttypes "github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
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
		if tc.expPanic {
			suite.Require().Panics(func() {
				clientState := ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, testClientHeight, commitmenttypes.GetSDKSpecs())
				suite.chainA.App.IBCKeeper.ClientKeeper.CreateClient(suite.chainA.GetContext(), tc.clientID, clientState, suite.consensusState)
			}, "Msg %d didn't panic: %s", i, tc.msg)
		} else {
			clientState := ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, testClientHeight, commitmenttypes.GetSDKSpecs())
			if tc.expPass {
				suite.Require().NotNil(clientState, "valid test case %d failed: %s", i, tc.msg)
			}
			// If we were able to NewClientState clientstate successfully, try persisting it to state
			_, err := suite.chainA.App.IBCKeeper.ClientKeeper.CreateClient(suite.chainA.GetContext(), tc.clientID, clientState, suite.consensusState)

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
	createFutureUpdateFn := func(s *KeeperTestSuite) ibctmtypes.Header {
		return ibctmtypes.CreateTestHeader(suite.chainA.ChainID,, int64(suite.header.GetHeight()+3), int64(suite.header.GetHeight()), suite.header.Header.Time.Add(time.Hour),
			suite.valSet, suite.valSet, []tmtypes.PrivValidator{suite.privVal})
	}
	createPastUpdateFn := func(s *KeeperTestSuite) ibctmtypes.Header {
		return ibctmtypes.CreateTestHeader(suite.chainA.ChainID,, int64(suite.header.GetHeight()-2), int64(suite.header.GetHeight())-4, suite.header.Header.Time,
			suite.valSet, suite.valSet, []tmtypes.PrivValidator{suite.privVal})
	}
	var (
		updateHeader ibctmtypes.Header
		clientState  *ibctmtypes.ClientState
	)

	cases := []struct {
		name     string
		malleate func() error
		expPass  bool
	}{
		{"valid update", func() error {
			clientState = ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, testClientHeight, commitmenttypes.GetSDKSpecs())
			_, err := suite.chainA.App.IBCKeeper.ClientKeeper.CreateClient(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState, suite.consensusState)

			// store intermediate consensus state to check that trustedHeight does not need to be highest consensus state before header height
			intermediateConsState := &ibctmtypes.ConsensusState{
				Height:             testClientHeight + 1,
				Timestamp:          suite.now.Add(time.Minute),
				NextValidatorsHash: suite.valSetHash,
			}
			suite.chainA.App.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], testClientHeight+1, intermediateConsState)

			clientState.LatestHeight = testClientHeight + 1
			suite.chainA.App.IBCKeeper.ClientKeeper.SetClientState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState)

			updateHeader = createFutureUpdateFn(suite)
			return err
		}, true},
		{"valid past update", func() error {
			clientState = ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, testClientHeight, commitmenttypes.GetSDKSpecs())
			_, err := suite.chainA.App.IBCKeeper.ClientKeeper.CreateClient(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState, suite.consensusState)
			suite.Require().NoError(err)

			// store previous consensus state
			prevConsState := &ibctmtypes.ConsensusState{
				Height:             1,
				Timestamp:          suite.past,
				NextValidatorsHash: suite.valSetHash,
			}
			suite.chainA.App.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], 1, prevConsState)

			// store intermediate consensus state to check that trustedHeight does not need to be hightest consensus state before header height
			intermediateConsState := &ibctmtypes.ConsensusState{
				Height:             2,
				Timestamp:          suite.past.Add(time.Minute),
				NextValidatorsHash: suite.valSetHash,
			}
			suite.chainA.App.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], 2, intermediateConsState)

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
			clientState = ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, testClientHeight, commitmenttypes.GetSDKSpecs())
			suite.chainA.App.IBCKeeper.ClientKeeper.SetClientState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState)
			suite.chainA.App.IBCKeeper.ClientKeeper.SetClientType(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], exported.Tendermint)
			updateHeader = createFutureUpdateFn(suite)

			return nil
		}, false},
		{"frozen client before update", func() error {
			clientState = &ibctmtypes.ClientState{FrozenHeight: 1, LatestHeight: testClientHeight}
			suite.chainA.App.IBCKeeper.ClientKeeper.SetClientState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState)
			suite.chainA.App.IBCKeeper.ClientKeeper.SetClientType(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], exported.Tendermint)
			updateHeader = createFutureUpdateFn(suite)

			return nil
		}, false},
		{"valid past update before client was frozen", func() error {
			clientState = ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, testClientHeight, commitmenttypes.GetSDKSpecs())
			clientState.FrozenHeight = testClientHeight - 1
			_, err := suite.chainA.App.IBCKeeper.ClientKeeper.CreateClient(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState, suite.consensusState)
			suite.Require().NoError(err)

			// store previous consensus state
			prevConsState := &ibctmtypes.ConsensusState{
				Height:             1,
				Timestamp:          suite.past,
				NextValidatorsHash: suite.valSetHash,
			}
			suite.chainA.App.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], 1, prevConsState)

			// updateHeader will fill in consensus state between prevConsState and suite.consState
			// clientState should not be updated
			updateHeader = createPastUpdateFn(suite)
			return nil
		}, true},
		{"invalid header", func() error {
			clientState = ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, testClientHeight, commitmenttypes.GetSDKSpecs())
			_, err := suite.chainA.App.IBCKeeper.ClientKeeper.CreateClient(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState, suite.consensusState)
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

			suite.ctx = suite.ctx.WithBlockTime(updateHeader.Header.Time.Add(time.Minute))

			updatedClientState, err := suite.chainA.App.IBCKeeper.ClientKeeper.UpdateClient(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], updateHeader)

			if tc.expPass {
				suite.Require().NoError(err, err)

				expConsensusState := &ibctmtypes.ConsensusState{
					Height:             updateHeader.GetHeight(),
					Timestamp:          updateHeader.GetTime(),
					Root:               commitmenttypes.NewMerkleRoot(updateHeader.Header.GetAppHash()),
					NextValidatorsHash: updateHeader.Header.NextValidatorsHash,
				}

				newClientState, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0])
				suite.Require().True(found, "valid test case %d failed: %s", i, tc.name)

				consensusState, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientConsensusState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], updateHeader.GetHeight())
				suite.Require().True(found, "valid test case %d failed: %s", i, tc.name)
				// check returned client state is same as client state in store
				suite.Require().Equal(updatedClientState, newClientState, "updatedClient state not persisted correctly")

				// Determine if clientState should be updated or not
				if uint64(updateHeader.GetHeight()) > clientState.GetLatestHeight() {
					// Header Height is greater than clientState latest Height, clientState should be updated with header.GetHeight()
					suite.Require().Equal(uint64(updateHeader.GetHeight()), updatedClientState.GetLatestHeight(), "clientstate height did not update")
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
	var localhostClient exported.ClientState = localhosttypes.NewClientState(suite.header.Header.GetChainID(), suite.chainA.GetContext().BlockHeight())

	suite.chainA.GetContext() = suite.chainA.GetContext().WithBlockHeight(suite.chainA.GetContext().BlockHeight() + 1)

	updatedClientState, err := suite.chainA.App.IBCKeeper.ClientKeeper.UpdateClient(suite.chainA.GetContext(), exported.ClientTypeLocalHost, nil)
	suite.Require().NoError(err, err)
	suite.Require().Equal(localhostClient.GetLatestHeight()+1, updatedClientState.GetLatestHeight())
}

func (suite *KeeperTestSuite) TestCheckMisbehaviourAndUpdateState() {
	altPrivVal := tmtypes.NewMockPV()
	altPubKey, err := altPrivVal.GetPubKey()
	suite.Require().NoError(err)
	altVal := tmtypes.NewValidator(altPubKey, 4)

	// Set valSet here with suite.valSet so it doesn't get reset on each testcase
	valSet := suite.valSet
	valsHash := valSet.Hash()

	// Create bothValSet with both suite validator and altVal
	bothValSet := tmtypes.NewValidatorSet(append(suite.valSet.Validators, altVal))
	bothValsHash := bothValSet.Hash()
	// Create alternative validator set with only altVal
	altValSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{altVal})

	// Create signer array and ensure it is in same order as bothValSet
	_, suiteVal := suite.valSet.GetByIndex(0)
	bothSigners := ibctmtypes.CreateSortedSignerArray(altPrivVal, suite.privVal, altVal, suiteVal)

	altSigners := []tmtypes.PrivValidator{altPrivVal}

	// Create valid Misbehaviour by making a duplicate header that signs over different block time
	altTime := suite.chainA.GetContext().BlockTime().Add(time.Minute)

	testCases := []struct {
		name     string
		evidence *ibctmtypes.Evidence
		malleate func() error
		expPass  bool
	}{
		{
			"trusting period misbehavior should pass",
			&ibctmtypes.Evidence{
				Header1:  ibctmtypes.CreateTestHeader(suite.chainA.ChainID, testClientHeight, testClientHeight, altTime, bothValSet, bothValSet, bothSigners),
				Header2:  ibctmtypes.CreateTestHeader(suite.chainA.ChainID, testClientHeight, testClientHeight, suite.chainA.GetContext().BlockTime(), bothValSet, bothValSet, bothSigners),
				ChainId:  suite.chainA.ChainID,
				ClientId: suite.chainA.ClientIDs[0],
			},
			func() error {
				suite.consensusState.NextValidatorsHash = bothValsHash
				clientState := ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, testClientHeight, commitmenttypes.GetSDKSpecs())
				_, err := suite.chainA.App.IBCKeeper.ClientKeeper.CreateClient(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState, suite.consensusState)

				return err
			},
			true,
		},
		{
			"misbehavior at later height should pass",
			&ibctmtypes.Evidence{
				Header1:  ibctmtypes.CreateTestHeader(suite.chainA.ChainID, testClientHeight+5, testClientHeight, altTime, bothValSet, valSet, bothSigners),
				Header2:  ibctmtypes.CreateTestHeader(suite.chainA.ChainID, testClientHeight+5, testClientHeight, suite.chainA.GetContext().BlockTime(), bothValSet, valSet, bothSigners),
				ChainId:  suite.chainA.ChainID,
				ClientId: suite.chainA.ClientIDs[0],
			},
			func() error {
				suite.consensusState.NextValidatorsHash = valsHash
				clientState := ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, testClientHeight, commitmenttypes.GetSDKSpecs())
				_, err := suite.chainA.App.IBCKeeper.ClientKeeper.CreateClient(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState, suite.consensusState)

				// store intermediate consensus state to check that trustedHeight does not need to be highest consensus state before header height
				intermediateConsState := &ibctmtypes.ConsensusState{
					Height:             testClientHeight + 3,
					Timestamp:          suite.now.Add(time.Minute),
					NextValidatorsHash: suite.valSetHash,
				}
				suite.chainA.App.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], testClientHeight+3, intermediateConsState)

				clientState.LatestHeight = testClientHeight + 3
				suite.chainA.App.IBCKeeper.ClientKeeper.SetClientState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState)

				return err
			},
			true,
		},
		{
			"misbehavior at later height with different trusted heights should pass",
			&ibctmtypes.Evidence{
				Header1:  ibctmtypes.CreateTestHeader(suite.chainA.ChainID,, testClientHeight+5, testClientHeight, altTime, bothValSet, valSet, bothSigners),
				Header2:  ibctmtypes.CreateTestHeader(suite.chainA.ChainID,, testClientHeight+5, testClientHeight+3, suite.chainA.GetContext().BlockTime(), bothValSet, bothValSet, bothSigners),
				ChainId:  suite.chainA.ChainID,,
				ClientId: suite.chainA.ClientIDs[0],
			},
			func() error {
				suite.consensusState.NextValidatorsHash = valsHash
				clientState := ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, testClientHeight, commitmenttypes.GetSDKSpecs())
				_, err := suite.chainA.App.IBCKeeper.ClientKeeper.CreateClient(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState, suite.consensusState)

				// store trusted consensus state for Header2
				intermediateConsState := &ibctmtypes.ConsensusState{
					Height:             testClientHeight + 3,
					Timestamp:          suite.now.Add(time.Minute),
					NextValidatorsHash: bothValsHash,
				}
				suite.chainA.App.IBCKeeper.ClientKeeper.SetClientConsensusState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], testClientHeight+3, intermediateConsState)

				clientState.LatestHeight = testClientHeight + 3
				suite.chainA.App.IBCKeeper.ClientKeeper.SetClientState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState)

				return err
			},
			true,
		},
		{
			"misbehaviour fails validatebasic",
			&ibctmtypes.Evidence{
				Header1:  ibctmtypes.CreateTestHeader(suite.chainA.ChainID,, testClientHeight+1, testClientHeight, altTime, bothValSet, bothValSet, bothSigners),
				Header2:  ibctmtypes.CreateTestHeader(suite.chainA.ChainID,, testClientHeight, testClientHeight, suite.chainA.GetContext().BlockTime(), bothValSet, bothValSet, bothSigners),
				ChainId:  suite.chainA.ChainID,,
				ClientId: suite.chainA.ClientIDs[0],
			},
			func() error {
				suite.consensusState.NextValidatorsHash = bothValsHash
				clientState := ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, testClientHeight, commitmenttypes.GetSDKSpecs())
				_, err := suite.chainA.App.IBCKeeper.ClientKeeper.CreateClient(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState, suite.consensusState)

				return err
			},
			false,
		},
		{
			"trusted ConsensusState1 not found",
			&ibctmtypes.Evidence{
				Header1:  ibctmtypes.CreateTestHeader(suite.chainA.ChainID,, testClientHeight+5, testClientHeight+3, altTime, bothValSet, bothValSet, bothSigners),
				Header2:  ibctmtypes.CreateTestHeader(suite.chainA.ChainID,, testClientHeight+5, testClientHeight, suite.chainA.GetContext().BlockTime(), bothValSet, valSet, bothSigners),
				ChainId:  suite.chainA.ChainID,,
				ClientId: suite.chainA.ClientIDs[0],
			},
			func() error {
				suite.consensusState.NextValidatorsHash = valsHash
				clientState := ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, testClientHeight, commitmenttypes.GetSDKSpecs())
				_, err := suite.chainA.App.IBCKeeper.ClientKeeper.CreateClient(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState, suite.consensusState)
				// intermediate consensus state at height + 3 is not created
				return err
			},
			false,
		},
		{
			"trusted ConsensusState2 not found",
			&ibctmtypes.Evidence{
				Header1:  ibctmtypes.CreateTestHeader(suite.chainA.ChainID,, testClientHeight+5, testClientHeight, altTime, bothValSet, valSet, bothSigners),
				Header2:  ibctmtypes.CreateTestHeader(suite.chainA.ChainID,, testClientHeight+5, testClientHeight+3, suite.chainA.GetContext().BlockTime(), bothValSet, bothValSet, bothSigners),
				ChainId:  suite.chainA.ChainID,,
				ClientId: suite.chainA.ClientIDs[0],
			},
			func() error {
				suite.consensusState.NextValidatorsHash = valsHash
				clientState := ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, testClientHeight, commitmenttypes.GetSDKSpecs())
				_, err := suite.chainA.App.IBCKeeper.ClientKeeper.CreateClient(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState, suite.consensusState)
				// intermediate consensus state at height + 3 is not created
				return err
			},
			false,
		},

		{
			"client state not found",
			&ibctmtypes.Evidence{},
			func() error { return nil },
			false,
		},
		{
			"client already frozen at earlier height",
			&ibctmtypes.Evidence{
				Header1:  ibctmtypes.CreateTestHeader(suite.chainA.ChainID,, testClientHeight, testClientHeight, altTime, bothValSet, bothValSet, bothSigners),
				Header2:  ibctmtypes.CreateTestHeader(suite.chainA.ChainID,, testClientHeight, testClientHeight, suite.chainA.GetContext().BlockTime(), bothValSet, bothValSet, bothSigners),
				ChainId:  suite.chainA.ChainID,,
				ClientId: suite.chainA.ClientIDs[0],
			},
			func() error {
				suite.consensusState.NextValidatorsHash = bothValsHash
				clientState := ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, testClientHeight, commitmenttypes.GetSDKSpecs())
				_, err := suite.chainA.App.IBCKeeper.ClientKeeper.CreateClient(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState, suite.consensusState)

				clientState.FrozenHeight = 1
				suite.chainA.App.IBCKeeper.ClientKeeper.SetClientState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState)

				return err
			},
			false,
		},

		{
			"misbehaviour check failed",
			&ibctmtypes.Evidence{
				Header1:  ibctmtypes.CreateTestHeader(suite.chainA.ChainID,, testClientHeight, testClientHeight, altTime, bothValSet, bothValSet, bothSigners),
				Header2:  ibctmtypes.CreateTestHeader(suite.chainA.ChainID,, testClientHeight, testClientHeight, suite.chainA.GetContext().BlockTime(), altValSet, bothValSet, altSigners),
				ChainId:  suite.chainA.ChainID,,
				ClientId: suite.chainA.ClientIDs[0],
			},
			func() error {
				clientState := ibctmtypes.NewClientState(suite.chainA.ChainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, testClientHeight, commitmenttypes.GetSDKSpecs())
				if err != nil {
					return err
				}
				_, err = suite.chainA.App.IBCKeeper.ClientKeeper.CreateClient(suite.chainA.GetContext(), suite.chainA.ClientIDs[0], clientState, suite.consensusState)

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

			err = suite.chainA.App.IBCKeeper.ClientKeeper.CheckMisbehaviourAndUpdateState(suite.chainA.GetContext(), tc.evidence)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)

				clientState, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientState(suite.chainA.GetContext(), suite.chainA.ClientIDs[0])
				suite.Require().True(found, "valid test case %d failed: %s", i, tc.name)
				suite.Require().True(clientState.IsFrozen(), "valid test case %d failed: %s", i, tc.name)
				suite.Require().Equal(uint64(tc.evidence.GetHeight()), clientState.GetFrozenHeight(),
					"valid test case %d failed: %s. Expected FrozenHeight %d got %d", tc.evidence.GetHeight(), clientState.GetFrozenHeight())
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
			}
		})
	}
}
