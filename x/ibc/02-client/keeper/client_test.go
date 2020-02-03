package keeper_test

import (
	"bytes"
	"fmt"

	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint"
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
		params   params
		expPass  bool
		expPanic bool
	}{
		{"success", params{testClientID, exported.Tendermint}, true, false},
		{"client ID exists", params{testClientID, exported.Tendermint}, false, false},
		{"client type exists", params{testClientID2, exported.Tendermint}, false, true},
		{"invalid client type", params{testClientID3, invalidClientType}, false, false},
	}

	for i, tc := range cases {
		tc := tc

		if tc.expPanic {
			suite.Require().Panics(func() {
				suite.keeper.CreateClient(suite.ctx, tc.params.clientID, tc.params.clientType, suite.consensusState)
			}, "Msg %d didn't panic: %s", i, tc.msg)
		} else {
			clientState, err := suite.keeper.CreateClient(suite.ctx, tc.params.clientID, tc.params.clientType, suite.consensusState)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
				suite.Require().NotNil(clientState, "valid test case %d failed: %s", i, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
				suite.Require().Nil(clientState, "invalid test case %d passed: %s", i, tc.msg)
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
			_, err := suite.keeper.CreateClient(suite.ctx, testClientID, exported.Tendermint, suite.consensusState)
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
			clientState := tendermint.ClientState{FrozenHeight: 1, ID: testClientID, LatestHeight: 10}
			suite.keeper.SetClientState(suite.ctx, clientState)
			suite.keeper.SetClientType(suite.ctx, testClientID, exported.Tendermint)
			return nil
		}, false},
		{"invalid header", func() error {
			_, err := suite.keeper.CreateClient(suite.ctx, testClientID, exported.Tendermint, suite.consensusState)
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
				expConsensusState := tendermint.ConsensusState{
					Timestamp:    suite.header.Time,
					Root:         commitment.NewRoot(suite.header.AppHash),
					ValidatorSet: suite.header.ValidatorSet,
				}

				clientState, found := suite.keeper.GetClientState(suite.ctx, testClientID)
				suite.Require().True(found, "valid test case %d failed: %s", i, tc.name)

				consensusState, found := suite.keeper.GetClientConsensusState(suite.ctx, testClientID, uint64(suite.header.GetHeight()))
				suite.Require().True(found, "valid test case %d failed: %s", i, tc.name)

				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
				suite.Require().Equal(suite.header.GetHeight(), clientState.GetLatestHeight(), "client state height not updated correctly")
				suite.Require().Equal(expConsensusState, consensusState, "consensus state should have been updated")
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
		evidence tendermint.Evidence
		malleate func() error
		expPass  bool
	}{
		{
			"trusting period misbehavior should pass",
			tendermint.Evidence{
				Header1:  tendermint.CreateTestHeader(testClientID, testClientHeight, suite.ctx.BlockTime(), bothValSet, suite.valSet, bothSigners),
				Header2:  tendermint.CreateTestHeader(testClientID, testClientHeight, suite.ctx.BlockTime(), bothValSet, bothValSet, bothSigners),
				ChainID:  testClientID,
				ClientID: testClientID,
			},
			func() error {
				suite.consensusState.ValidatorSet = bothValSet
				_, err := suite.keeper.CreateClient(suite.ctx, testClientID, exported.Tendermint, suite.consensusState)
				return err
			},
			true,
		},
		{
			"client state not found",
			tendermint.Evidence{},
			func() error { return nil },
			false,
		},
		{
			"consensus state not found",
			tendermint.Evidence{
				Header1:  tendermint.CreateTestHeader(testClientID, testClientHeight, suite.ctx.BlockTime(), bothValSet, suite.valSet, bothSigners),
				Header2:  tendermint.CreateTestHeader(testClientID, testClientHeight, suite.ctx.BlockTime(), bothValSet, bothValSet, bothSigners),
				ChainID:  testClientID,
				ClientID: testClientID,
			},
			func() error {
				clientState := tendermint.ClientState{FrozenHeight: 1, ID: testClientID, LatestHeight: 10}
				suite.keeper.SetClientState(suite.ctx, clientState)
				return nil
			},
			false,
		},
		{
			"consensus state not found",
			tendermint.Evidence{
				Header1:  tendermint.CreateTestHeader(testClientID, testClientHeight, suite.ctx.BlockTime(), bothValSet, suite.valSet, bothSigners),
				Header2:  tendermint.CreateTestHeader(testClientID, testClientHeight, suite.ctx.BlockTime(), bothValSet, bothValSet, bothSigners),
				ChainID:  testClientID,
				ClientID: testClientID,
			},
			func() error {
				clientState := tendermint.ClientState{FrozenHeight: 1, ID: testClientID, LatestHeight: 10}
				suite.keeper.SetClientState(suite.ctx, clientState)
				return nil
			},
			false,
		},
		{
			"misbehaviour check failed",
			tendermint.Evidence{
				Header1:  tendermint.CreateTestHeader(testClientID, testClientHeight, suite.ctx.BlockTime(), bothValSet, bothValSet, bothSigners),
				Header2:  tendermint.CreateTestHeader(testClientID, testClientHeight, suite.ctx.BlockTime(), altValSet, bothValSet, altSigners),
				ChainID:  testClientID,
				ClientID: testClientID,
			},
			func() error {
				_, err := suite.keeper.CreateClient(suite.ctx, testClientID, exported.Tendermint, suite.consensusState)
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
