package keeper_test

import (
	"bytes"
	"fmt"
	"time"

	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	localhosttypes "github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/types"

	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

const (
	invalidClientType exported.ClientType = 0
)

func (suite *KeeperTestSuite) TestCreateClient() {
	suite.keeper.SetClientType(suite.ctx, testClientID2, exported.Tendermint)

	cases := []struct {
		msg      string
		clientID string
		expPass  bool
		expPanic bool
	}{
		{"success", testClientID, true, false},
		{"client ID exists", testClientID, false, false},
		{"client type exists", testClientID2, false, true},
	}

	for i, tc := range cases {
		tc := tc
		i := i
		if tc.expPanic {
			suite.Require().Panics(func() {
				clientState := ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, testClientHeight, commitmenttypes.GetSDKSpecs())
				suite.keeper.CreateClient(suite.ctx, tc.clientID, clientState, suite.consensusState)
			}, "Msg %d didn't panic: %s", i, tc.msg)
		} else {
			clientState := ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, testClientHeight, commitmenttypes.GetSDKSpecs())
			if tc.expPass {
				suite.Require().NotNil(clientState, "valid test case %d failed: %s", i, tc.msg)
			}
			// If we were able to NewClientState clientstate successfully, try persisting it to state
			_, err := suite.keeper.CreateClient(suite.ctx, tc.clientID, clientState, suite.consensusState)

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
		return ibctmtypes.CreateTestHeader(testChainID, suite.header.Height+1, suite.header.Time.Add(time.Minute),
			suite.valSet, []tmtypes.PrivValidator{suite.privVal})
	}
	createPastUpdateFn := func(s *KeeperTestSuite) ibctmtypes.Header {
		return ibctmtypes.CreateTestHeader(testChainID, suite.header.Height-3, suite.header.Time,
			suite.valSet, []tmtypes.PrivValidator{suite.privVal})
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
			clientState = ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, testClientHeight, commitmenttypes.GetSDKSpecs())
			_, err := suite.keeper.CreateClient(suite.ctx, testClientID, clientState, suite.consensusState)
			updateHeader = createFutureUpdateFn(suite)
			return err
		}, true},
		{"valid past update", func() error {
			clientState = ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, testClientHeight, commitmenttypes.GetSDKSpecs())
			_, err := suite.keeper.CreateClient(suite.ctx, testClientID, clientState, suite.consensusState)
			suite.Require().NoError(err)

			// store previous consensus state
			prevConsState := ibctmtypes.ConsensusState{
				Height:             1,
				Timestamp:          suite.past,
				NextValidatorsHash: suite.valSet.Hash(),
				ValidatorSet:       suite.valSet,
			}
			suite.keeper.SetClientConsensusState(suite.ctx, testClientID, 1, prevConsState)

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
			suite.keeper.SetClientType(suite.ctx, testClientID, invalidClientType)
			updateHeader = createFutureUpdateFn(suite)

			return nil
		}, false},
		{"client state not found", func() error {
			suite.keeper.SetClientType(suite.ctx, testClientID, exported.Tendermint)
			updateHeader = createFutureUpdateFn(suite)

			return nil
		}, false},
		{"consensus state not found", func() error {
			clientState = ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, testClientHeight, commitmenttypes.GetSDKSpecs())
			suite.keeper.SetClientState(suite.ctx, testClientID, clientState)
			suite.keeper.SetClientType(suite.ctx, testClientID, exported.Tendermint)
			updateHeader = createFutureUpdateFn(suite)

			return nil
		}, false},
		{"frozen client before update", func() error {
			clientState = &ibctmtypes.ClientState{FrozenHeight: 1, LatestHeight: testClientHeight}
			suite.keeper.SetClientState(suite.ctx, testClientID, clientState)
			suite.keeper.SetClientType(suite.ctx, testClientID, exported.Tendermint)
			updateHeader = createFutureUpdateFn(suite)

			return nil
		}, false},
		{"valid past update before client was frozen", func() error {
			clientState = ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, testClientHeight, commitmenttypes.GetSDKSpecs())
			clientState.FrozenHeight = testClientHeight - 1
			_, err := suite.keeper.CreateClient(suite.ctx, testClientID, clientState, suite.consensusState)
			suite.Require().NoError(err)

			// store previous consensus state
			prevConsState := ibctmtypes.ConsensusState{
				Height:             1,
				Timestamp:          suite.past,
				NextValidatorsHash: suite.valSet.Hash(),
				ValidatorSet:       suite.valSet,
			}
			suite.keeper.SetClientConsensusState(suite.ctx, testClientID, 1, prevConsState)

			// updateHeader will fill in consensus state between prevConsState and suite.consState
			// clientState should not be updated
			updateHeader = createPastUpdateFn(suite)
			return nil
		}, true},
		{"invalid header", func() error {
			clientState = ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, testClientHeight, commitmenttypes.GetSDKSpecs())
			_, err := suite.keeper.CreateClient(suite.ctx, testClientID, clientState, suite.consensusState)
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

			suite.ctx = suite.ctx.WithBlockTime(updateHeader.Time.Add(time.Minute))

			updatedClientState, err := suite.keeper.UpdateClient(suite.ctx, testClientID, updateHeader)

			if tc.expPass {
				suite.Require().NoError(err, err)

				expConsensusState := ibctmtypes.ConsensusState{
					Height:             updateHeader.GetHeight(),
					Timestamp:          updateHeader.Time,
					Root:               commitmenttypes.NewMerkleRoot(updateHeader.AppHash),
					NextValidatorsHash: updateHeader.NextValidatorsHash,
					ValidatorSet:       updateHeader.ValidatorSet,
				}

				newClientState, found := suite.keeper.GetClientState(suite.ctx, testClientID)
				suite.Require().True(found, "valid test case %d failed: %s", i, tc.name)

				consensusState, found := suite.keeper.GetClientConsensusState(suite.ctx, testClientID, updateHeader.GetHeight())
				suite.Require().True(found, "valid test case %d failed: %s", i, tc.name)
				tmConsState, ok := consensusState.(ibctmtypes.ConsensusState)
				suite.Require().True(ok, "consensus state is not a tendermint consensus state")
				// recalculate cached totalVotingPower field for equality check
				tmConsState.ValidatorSet.TotalVotingPower()

				// check returned client state is same as client state in store
				suite.Require().Equal(updatedClientState, newClientState, "updatedClient state not persisted correctly")

				// Determine if clientState should be updated or not
				if uint64(updateHeader.Height) > clientState.GetLatestHeight() {
					// Header Height is greater than clientState latest Height, clientState should be updated with header.Height
					suite.Require().Equal(uint64(updateHeader.Height), updatedClientState.GetLatestHeight(), "clientstate height did not update")
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
	var localhostClient exported.ClientState = localhosttypes.NewClientState(suite.header.ChainID, suite.ctx.BlockHeight())

	suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)

	updatedClientState, err := suite.keeper.UpdateClient(suite.ctx, exported.ClientTypeLocalHost, nil)
	suite.Require().NoError(err, err)
	suite.Require().Equal(localhostClient.GetLatestHeight()+1, updatedClientState.GetLatestHeight())
}

func (suite *KeeperTestSuite) TestCheckMisbehaviourAndUpdateState() {
	altPrivVal := tmtypes.NewMockPV()
	altPubKey, err := altPrivVal.GetPubKey()
	suite.Require().NoError(err)
	altVal := tmtypes.NewValidator(altPubKey, 4)

	// Create bothValSet with both suite validator and altVal
	bothValSet := tmtypes.NewValidatorSet(append(suite.valSet.Validators, altVal))
	// Create alternative validator set with only altVal
	altValSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{altVal})

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
		name     string
		evidence ibctmtypes.Evidence
		malleate func() error
		expPass  bool
	}{
		{
			"trusting period misbehavior should pass",
			ibctmtypes.Evidence{
				Header1:  ibctmtypes.CreateTestHeader(testClientID, testClientHeight, suite.ctx.BlockTime(), bothValSet, bothSigners),
				Header2:  ibctmtypes.CreateTestHeader(testClientID, testClientHeight, suite.ctx.BlockTime(), bothValSet, bothSigners),
				ChainID:  testClientID,
				ClientID: testClientID,
			},
			func() error {
				suite.consensusState.ValidatorSet = bothValSet
				clientState := ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, testClientHeight, commitmenttypes.GetSDKSpecs())
				_, err := suite.keeper.CreateClient(suite.ctx, testClientID, clientState, suite.consensusState)

				return err
			},
			true,
		},
		{
			"misbehavior at later height should pass",
			ibctmtypes.Evidence{
				Header1:  ibctmtypes.CreateTestHeader(testClientID, testClientHeight+5, suite.ctx.BlockTime(), bothValSet, bothSigners),
				Header2:  ibctmtypes.CreateTestHeader(testClientID, testClientHeight+5, suite.ctx.BlockTime(), bothValSet, bothSigners),
				ChainID:  testClientID,
				ClientID: testClientID,
			},
			func() error {
				suite.consensusState.ValidatorSet = bothValSet
				clientState := ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, testClientHeight, commitmenttypes.GetSDKSpecs())
				_, err := suite.keeper.CreateClient(suite.ctx, testClientID, clientState, suite.consensusState)

				return err
			},
			true,
		},
		{
			"client state not found",
			ibctmtypes.Evidence{},
			func() error { return nil },
			false,
		},
		{
			"consensus state not found",
			ibctmtypes.Evidence{
				Header1:  ibctmtypes.CreateTestHeader(testClientID, testClientHeight, suite.ctx.BlockTime(), bothValSet, bothSigners),
				Header2:  ibctmtypes.CreateTestHeader(testClientID, testClientHeight, suite.ctx.BlockTime(), bothValSet, bothSigners),
				ChainID:  testClientID,
				ClientID: testClientID,
			},
			func() error {
				clientState := &ibctmtypes.ClientState{FrozenHeight: 1, LatestHeight: testClientHeight}
				suite.keeper.SetClientState(suite.ctx, testClientID, clientState)
				return nil
			},
			false,
		},
		{
			"consensus state not found",
			ibctmtypes.Evidence{
				Header1:  ibctmtypes.CreateTestHeader(testClientID, testClientHeight, suite.ctx.BlockTime(), bothValSet, bothSigners),
				Header2:  ibctmtypes.CreateTestHeader(testClientID, testClientHeight, suite.ctx.BlockTime(), bothValSet, bothSigners),
				ChainID:  testClientID,
				ClientID: testClientID,
			},
			func() error {
				clientState := &ibctmtypes.ClientState{FrozenHeight: 1, LatestHeight: testClientHeight}
				suite.keeper.SetClientState(suite.ctx, testClientID, clientState)
				return nil
			},
			false,
		},
		{
			"misbehaviour check failed",
			ibctmtypes.Evidence{
				Header1:  ibctmtypes.CreateTestHeader(testClientID, testClientHeight, suite.ctx.BlockTime(), bothValSet, bothSigners),
				Header2:  ibctmtypes.CreateTestHeader(testClientID, testClientHeight, suite.ctx.BlockTime(), altValSet, altSigners),
				ChainID:  testClientID,
				ClientID: testClientID,
			},
			func() error {
				clientState := ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, testClientHeight, commitmenttypes.GetSDKSpecs())
				if err != nil {
					return err
				}
				_, err = suite.keeper.CreateClient(suite.ctx, testClientID, clientState, suite.consensusState)

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

			err = suite.keeper.CheckMisbehaviourAndUpdateState(suite.ctx, tc.evidence)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)

				clientState, found := suite.keeper.GetClientState(suite.ctx, testClientID)
				suite.Require().True(found, "valid test case %d failed: %s", i, tc.name)
				suite.Require().True(clientState.IsFrozen(), "valid test case %d failed: %s", i, tc.name)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
			}
		})
	}
}
