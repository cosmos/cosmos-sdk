package keeper_test

import (
	"bytes"
	"fmt"
	"time"

	lite "github.com/tendermint/tendermint/lite2"
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
				clientState, err := ibctmtypes.Initialize(tc.clientID, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, suite.header, commitmenttypes.GetSDKSpecs())
				suite.Require().NoError(err, "err on client state initialization")
				suite.keeper.CreateClient(suite.ctx, clientState, suite.consensusState)
			}, "Msg %d didn't panic: %s", i, tc.msg)
		} else {
			clientState, err := ibctmtypes.Initialize(tc.clientID, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, suite.header, commitmenttypes.GetSDKSpecs())
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

func (suite *KeeperTestSuite) TestUpdateClientTendermint() {
	// Must create header creation functions since suite.header gets recreated on each test case
	createValidUpdateFn := func(s *KeeperTestSuite) ibctmtypes.Header {
		return ibctmtypes.CreateTestHeader(testClientID, exported.NewHeight(0, suite.header.Height.EpochHeight+1), suite.header.Time.Add(time.Minute),
			suite.valSet, []tmtypes.PrivValidator{suite.privVal})
	}
	createInvalidUpdateFn := func(s *KeeperTestSuite) ibctmtypes.Header {
		return ibctmtypes.CreateTestHeader(testClientID, exported.NewHeight(0, suite.header.Height.EpochHeight-3), suite.header.Time.Add(time.Minute),
			suite.valSet, []tmtypes.PrivValidator{suite.privVal})
	}
	createNextEpochHeaderFn := func(s *KeeperTestSuite, height exported.Height) ibctmtypes.Header {
		return ibctmtypes.CreateTestHeader(testClientID, height, suite.header.Time.Add(time.Minute), suite.valSet, []tmtypes.PrivValidator{suite.privVal})
	}
	var updateHeader ibctmtypes.Header

	cases := []struct {
		name     string
		malleate func() error
		expPass  bool
	}{
		{"valid update", func() error {
			clientState, err := ibctmtypes.Initialize(testClientID, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, suite.header, commitmenttypes.GetSDKSpecs())
			if err != nil {
				return err
			}
			_, err = suite.keeper.CreateClient(suite.ctx, clientState, suite.consensusState)
			updateHeader = createValidUpdateFn(suite)
			return err
		}, true},
		{"client type not found", func() error {
			updateHeader = createValidUpdateFn(suite)

			return nil
		}, false},
		{"client type and header type mismatch", func() error {
			suite.keeper.SetClientType(suite.ctx, testClientID, invalidClientType)
			updateHeader = createValidUpdateFn(suite)

			return nil
		}, false},
		{"client state not found", func() error {
			suite.keeper.SetClientType(suite.ctx, testClientID, exported.Tendermint)
			updateHeader = createValidUpdateFn(suite)

			return nil
		}, false},
		{"frozen client", func() error {
			clientState := ibctmtypes.ClientState{FrozenHeight: exported.NewHeight(0, 1), ID: testClientID, LastHeader: suite.header}
			suite.keeper.SetClientState(suite.ctx, clientState)
			suite.keeper.SetClientType(suite.ctx, testClientID, exported.Tendermint)
			updateHeader = createValidUpdateFn(suite)

			return nil
		}, false},
		{"height of earlier epoch froze", func() error {
			currentHeight := exported.NewHeight(3, 4)
			frozenHeight := exported.NewHeight(2, 5)
			updateHeight := exported.NewHeight(3, 8)
			clientState := ibctmtypes.ClientState{FrozenHeight: frozenHeight, ID: testClientID, LastHeader: createNextEpochHeaderFn(suite, currentHeight)}
			suite.keeper.SetClientState(suite.ctx, clientState)
			suite.keeper.SetClientType(suite.ctx, testClientID, exported.Tendermint)
			updateHeader = createNextEpochHeaderFn(suite, updateHeight)

			return nil
		}, false},
		{"invalid header", func() error {
			clientState, err := ibctmtypes.Initialize(testClientID, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, suite.header, commitmenttypes.GetSDKSpecs())
			if err != nil {
				return err
			}
			_, err = suite.keeper.CreateClient(suite.ctx, clientState, suite.consensusState)
			if err != nil {
				return err
			}
			updateHeader = createInvalidUpdateFn(suite)

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
					Height:       updateHeader.GetHeight(),
					Timestamp:    updateHeader.Time,
					Root:         commitmenttypes.NewMerkleRoot(updateHeader.AppHash),
					ValidatorSet: updateHeader.ValidatorSet,
				}

				clientState, found := suite.keeper.GetClientState(suite.ctx, testClientID)
				suite.Require().True(found, "valid test case %d failed: %s", i, tc.name)

				consensusState, found := suite.keeper.GetClientConsensusState(suite.ctx, testClientID, updateHeader.GetHeight())
				suite.Require().True(found, "valid test case %d failed: %s", i, tc.name)
				tmConsState, ok := consensusState.(ibctmtypes.ConsensusState)
				suite.Require().True(ok, "consensus state is not a tendermint consensus state")
				// recalculate cached totalVotingPower field for equality check
				tmConsState.ValidatorSet.TotalVotingPower()

				tmClientState, ok := updatedClientState.(ibctmtypes.ClientState)
				suite.Require().True(ok, "client state is not a tendermint client state")

				// recalculate cached totalVotingPower field for equality check
				tmClientState.LastHeader.ValidatorSet.TotalVotingPower()

				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
				suite.Require().Equal(updateHeader.GetHeight(), clientState.GetLatestHeight(), "client state height not updated correctly on case %s", tc.name)
				suite.Require().Equal(expConsensusState, consensusState, "consensus state should have been updated on case %s", tc.name)
				suite.Require().Equal(updatedClientState, tmClientState, "client states don't match")
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
	suite.Require().Equal(localhostClient.GetID(), updatedClientState.GetID())
	suite.Require().Equal(localhostClient.GetLatestHeight().EpochNumber, updatedClientState.GetLatestHeight().EpochNumber)
	suite.Require().Equal(localhostClient.GetLatestHeight().EpochHeight+1, updatedClientState.GetLatestHeight().EpochHeight)
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
				clientState, err := ibctmtypes.Initialize(testClientID, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, suite.header, commitmenttypes.GetSDKSpecs())
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
				Header1:  ibctmtypes.CreateTestHeader(testClientID, exported.NewHeight(0, testClientHeight.EpochHeight+5), suite.ctx.BlockTime(), bothValSet, bothSigners),
				Header2:  ibctmtypes.CreateTestHeader(testClientID, exported.NewHeight(0, testClientHeight.EpochHeight+5), suite.ctx.BlockTime(), bothValSet, bothSigners),
				ChainID:  testClientID,
				ClientID: testClientID,
			},
			func() error {
				suite.consensusState.ValidatorSet = bothValSet
				clientState, err := ibctmtypes.Initialize(testClientID, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, suite.header, commitmenttypes.GetSDKSpecs())
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
				Header1:  ibctmtypes.CreateTestHeader(testClientID, testClientHeight, suite.ctx.BlockTime(), bothValSet, bothSigners),
				Header2:  ibctmtypes.CreateTestHeader(testClientID, testClientHeight, suite.ctx.BlockTime(), bothValSet, bothSigners),
				ChainID:  testClientID,
				ClientID: testClientID,
			},
			func() error {
				clientState := ibctmtypes.ClientState{FrozenHeight: exported.NewHeight(0, 1), ID: testClientID, LastHeader: suite.header}
				suite.keeper.SetClientState(suite.ctx, clientState)
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
				clientState := ibctmtypes.ClientState{FrozenHeight: exported.NewHeight(0, 1), ID: testClientID, LastHeader: suite.header}
				suite.keeper.SetClientState(suite.ctx, clientState)
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
				clientState, err := ibctmtypes.Initialize(testClientID, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, suite.header, commitmenttypes.GetSDKSpecs())
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
