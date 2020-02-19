package keeper_test

import (
	"bytes"
	"fmt"

	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"

	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

const (
	invalidClientType exported.ClientType = 0
)

func (suite *KeeperTestSuite) TestCreateClient() {
	type params struct {
		clientID   string
		clientType exported.ClientType
	}

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

		if tc.expPanic {
			suite.Require().Panics(func() {
				clientState, err := ibctmtypes.Initialize(tc.clientID, trustingPeriod, ubdPeriod, suite.header)
				suite.Require().NoError(err, "err on client state initialization")
				suite.keeper.CreateClient(suite.ctx, clientState, suite.consensusState)
			}, "Msg %d didn't panic: %s", i, tc.msg)
		} else {
			clientState, err := ibctmtypes.Initialize(tc.clientID, trustingPeriod, ubdPeriod, suite.header)
			if tc.expPass {
				suite.Require().NoError(err, "errored on initialization")
				suite.Require().NotNil(clientState, "valid test case %d failed: %s", i, tc.msg)
			}
			// If we were able to initialize clientstate successfully, try persisting it to state
			if err == nil {
				_, err = suite.keeper.CreateClient(suite.ctx, clientState, suite.consensusState)
			}

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		}
	}
}

func (suite *KeeperTestSuite) TestUpdateClient() {
	cases := []struct {
		name     string
		malleate func() error
		expPass  bool
	}{
		{"valid update", func() error {
			clientState, err := ibctmtypes.Initialize(testClientID, trustingPeriod, ubdPeriod, suite.header)
			if err != nil {
				return err
			}
			_, err = suite.keeper.CreateClient(suite.ctx, clientState, suite.consensusState)
			return err
		}, true},
		{"client type not found", func() error {
			return nil
		}, false},
		{"client type and header type mismatch", func() error {
			suite.keeper.SetClientType(suite.ctx, testClientID, invalidClientType)
			return nil
		}, false},
		{"client state not found", func() error {
			suite.keeper.SetClientType(suite.ctx, testClientID, exported.Tendermint)
			return nil
		}, false},
		{"frozen client", func() error {
			clientState := ibctmtypes.ClientState{FrozenHeight: 1, ID: testClientID, LastHeader: suite.header}
			suite.keeper.SetClientState(suite.ctx, clientState)
			suite.keeper.SetClientType(suite.ctx, testClientID, exported.Tendermint)
			return nil
		}, false},
		{"invalid header", func() error {
			clientState, err := ibctmtypes.Initialize(testClientID, trustingPeriod, ubdPeriod, suite.header)
			if err != nil {
				return err
			}
			_, err = suite.keeper.CreateClient(suite.ctx, clientState, suite.consensusState)
			if err != nil {
				return err
			}
			suite.header.Height = suite.ctx.BlockHeight() - 1
			return nil
		}, false},
	}

	for i, tc := range cases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()

			err := tc.malleate()
			suite.Require().NoError(err)

			err = suite.keeper.UpdateClient(suite.ctx, testClientID, suite.header)

			if tc.expPass {
				expConsensusState := ibctmtypes.ConsensusState{
					Timestamp:    suite.header.Time,
					Root:         commitment.NewRoot(suite.header.AppHash),
					ValidatorSet: suite.header.ValidatorSet,
				}

				clientState, found := suite.keeper.GetClientState(suite.ctx, testClientID)
				suite.Require().True(found, "valid test case %d failed: %s", i, tc.name)

				consensusState, found := suite.keeper.GetClientConsensusState(suite.ctx, testClientID, uint64(suite.header.GetHeight()))
				suite.Require().True(found, "valid test case %d failed: %s", i, tc.name)
				tmConsState, ok := consensusState.(ibctmtypes.ConsensusState)
				suite.Require().True(ok, "consensus state is not a tendermint consensus state")
				// recalculate cached totalVotingPower field for equality check
				tmConsState.ValidatorSet.TotalVotingPower()

				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
				suite.Require().Equal(suite.header.GetHeight(), clientState.GetLatestHeight(), "client state height not updated correctly on case %s", tc.name)
				suite.Require().Equal(expConsensusState, consensusState, "consensus state should have been updated on case %s", tc.name)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestCheckMisbehaviourAndUpdateState() {
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
		name     string
		evidence ibctmtypes.Evidence
		malleate func() error
		expPass  bool
	}{
		{
			"trusting period misbehavior should pass",
			ibctmtypes.Evidence{
				Header1:  ibctmtypes.CreateTestHeader(testClientID, testClientHeight, suite.ctx.BlockTime(), bothValSet, suite.valSet, bothSigners),
				Header2:  ibctmtypes.CreateTestHeader(testClientID, testClientHeight, suite.ctx.BlockTime(), bothValSet, bothValSet, bothSigners),
				ChainID:  testClientID,
				ClientID: testClientID,
			},
			func() error {
				suite.consensusState.ValidatorSet = bothValSet
				clientState, err := ibctmtypes.Initialize(testClientID, trustingPeriod, ubdPeriod, suite.header)
				if err != nil {
					return err
				}
				_, err = suite.keeper.CreateClient(suite.ctx, clientState, suite.consensusState)

				return err
			},
			true,
		},
		{
			"misbehavior at later height should pass",
			ibctmtypes.Evidence{
				Header1:  ibctmtypes.CreateTestHeader(testClientID, testClientHeight+5, suite.ctx.BlockTime(), bothValSet, suite.valSet, bothSigners),
				Header2:  ibctmtypes.CreateTestHeader(testClientID, testClientHeight+5, suite.ctx.BlockTime(), bothValSet, bothValSet, bothSigners),
				ChainID:  testClientID,
				ClientID: testClientID,
			},
			func() error {
				suite.consensusState.ValidatorSet = bothValSet
				clientState, err := ibctmtypes.Initialize(testClientID, trustingPeriod, ubdPeriod, suite.header)
				if err != nil {
					return err
				}
				_, err = suite.keeper.CreateClient(suite.ctx, clientState, suite.consensusState)

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
				Header1:  ibctmtypes.CreateTestHeader(testClientID, testClientHeight, suite.ctx.BlockTime(), bothValSet, suite.valSet, bothSigners),
				Header2:  ibctmtypes.CreateTestHeader(testClientID, testClientHeight, suite.ctx.BlockTime(), bothValSet, bothValSet, bothSigners),
				ChainID:  testClientID,
				ClientID: testClientID,
			},
			func() error {
				clientState := ibctmtypes.ClientState{FrozenHeight: 1, ID: testClientID, LastHeader: suite.header}
				suite.keeper.SetClientState(suite.ctx, clientState)
				return nil
			},
			false,
		},
		{
			"consensus state not found",
			ibctmtypes.Evidence{
				Header1:  ibctmtypes.CreateTestHeader(testClientID, testClientHeight, suite.ctx.BlockTime(), bothValSet, suite.valSet, bothSigners),
				Header2:  ibctmtypes.CreateTestHeader(testClientID, testClientHeight, suite.ctx.BlockTime(), bothValSet, bothValSet, bothSigners),
				ChainID:  testClientID,
				ClientID: testClientID,
			},
			func() error {
				clientState := ibctmtypes.ClientState{FrozenHeight: 1, ID: testClientID, LastHeader: suite.header}
				suite.keeper.SetClientState(suite.ctx, clientState)
				return nil
			},
			false,
		},
		{
			"misbehaviour check failed",
			ibctmtypes.Evidence{
				Header1:  ibctmtypes.CreateTestHeader(testClientID, testClientHeight, suite.ctx.BlockTime(), bothValSet, bothValSet, bothSigners),
				Header2:  ibctmtypes.CreateTestHeader(testClientID, testClientHeight, suite.ctx.BlockTime(), altValSet, bothValSet, altSigners),
				ChainID:  testClientID,
				ClientID: testClientID,
			},
			func() error {
				clientState, err := ibctmtypes.Initialize(testClientID, trustingPeriod, ubdPeriod, suite.header)
				if err != nil {
					return err
				}
				_, err = suite.keeper.CreateClient(suite.ctx, clientState, suite.consensusState)

				return err
			},
			false,
		},
	}

	for i, tc := range testCases {
		tc := tc
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
