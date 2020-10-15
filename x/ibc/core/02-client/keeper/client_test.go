package keeper_test

import (
	"fmt"
	"time"

	tmtypes "github.com/tendermint/tendermint/types"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/23-commitment/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
	localhosttypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/09-localhost/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
	ibctestingmock "github.com/cosmos/cosmos-sdk/x/ibc/testing/mock"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func (suite *KeeperTestSuite) TestCreateClient() {
	cases := []struct {
		msg      string
		clientID string
		expPass  bool
		expPanic bool
	}{
		{"success", testClientID, true, false},
		{"client ID exists", testClientID, false, false},
	}

	for i, tc := range cases {
		tc := tc
		i := i
		if tc.expPanic {
			suite.Require().Panics(func() {
				clientState := ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, testClientHeight, ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false)
				suite.keeper.CreateClient(suite.ctx, tc.clientID, clientState, suite.consensusState)
			}, "Msg %d didn't panic: %s", i, tc.msg)
		} else {
			clientState := ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, testClientHeight, ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false)
			if tc.expPass {
				suite.Require().NotNil(clientState, "valid test case %d failed: %s", i, tc.msg)
			}
			// If we were able to NewClientState clientstate successfully, try persisting it to state
			err := suite.keeper.CreateClient(suite.ctx, tc.clientID, clientState, suite.consensusState)

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
		heightPlus3 := clienttypes.NewHeight(suite.header.GetHeight().GetVersionNumber(), suite.header.GetHeight().GetVersionHeight()+3)
		height := suite.header.GetHeight().(clienttypes.Height)

		return suite.chainA.CreateTMClientHeader(testChainID, int64(heightPlus3.VersionHeight), height, suite.header.Header.Time.Add(time.Hour),
			suite.valSet, suite.valSet, []tmtypes.PrivValidator{suite.privVal})
	}
	createPastUpdateFn := func(s *KeeperTestSuite) *ibctmtypes.Header {
		heightMinus2 := clienttypes.NewHeight(suite.header.GetHeight().GetVersionNumber(), suite.header.GetHeight().GetVersionHeight()-2)
		heightMinus4 := clienttypes.NewHeight(suite.header.GetHeight().GetVersionNumber(), suite.header.GetHeight().GetVersionHeight()-4)

		return suite.chainA.CreateTMClientHeader(testChainID, int64(heightMinus2.VersionHeight), heightMinus4, suite.header.Header.Time,
			suite.valSet, suite.valSet, []tmtypes.PrivValidator{suite.privVal})
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
			clientState = ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, testClientHeight, ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false)
			err := suite.keeper.CreateClient(suite.ctx, testClientID, clientState, suite.consensusState)

			// store intermediate consensus state to check that trustedHeight does not need to be highest consensus state before header height
			incrementedClientHeight := testClientHeight.Increment()
			intermediateConsState := &ibctmtypes.ConsensusState{
				Timestamp:          suite.now.Add(time.Minute),
				NextValidatorsHash: suite.valSetHash,
			}
			suite.keeper.SetClientConsensusState(suite.ctx, testClientID, incrementedClientHeight, intermediateConsState)

			clientState.LatestHeight = incrementedClientHeight
			suite.keeper.SetClientState(suite.ctx, testClientID, clientState)

			updateHeader = createFutureUpdateFn(suite)
			return err
		}, true},
		{"valid past update", func() error {
			clientState = ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, testClientHeight, ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false)
			err := suite.keeper.CreateClient(suite.ctx, testClientID, clientState, suite.consensusState)
			suite.Require().NoError(err)

			height1 := types.NewHeight(0, 1)

			// store previous consensus state
			prevConsState := &ibctmtypes.ConsensusState{
				Timestamp:          suite.past,
				NextValidatorsHash: suite.valSetHash,
			}
			suite.keeper.SetClientConsensusState(suite.ctx, testClientID, height1, prevConsState)

			height2 := types.NewHeight(0, 2)

			// store intermediate consensus state to check that trustedHeight does not need to be hightest consensus state before header height
			intermediateConsState := &ibctmtypes.ConsensusState{
				Timestamp:          suite.past.Add(time.Minute),
				NextValidatorsHash: suite.valSetHash,
			}
			suite.keeper.SetClientConsensusState(suite.ctx, testClientID, height2, intermediateConsState)

			// updateHeader will fill in consensus state between prevConsState and suite.consState
			// clientState should not be updated
			updateHeader = createPastUpdateFn(suite)
			return nil
		}, true},
		{"client state not found", func() error {
			updateHeader = createFutureUpdateFn(suite)

			return nil
		}, false},
		{"consensus state not found", func() error {
			clientState = ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, testClientHeight, ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false)
			suite.keeper.SetClientState(suite.ctx, testClientID, clientState)
			updateHeader = createFutureUpdateFn(suite)

			return nil
		}, false},
		{"frozen client before update", func() error {
			clientState = &ibctmtypes.ClientState{FrozenHeight: types.NewHeight(0, 1), LatestHeight: testClientHeight}
			suite.keeper.SetClientState(suite.ctx, testClientID, clientState)
			updateHeader = createFutureUpdateFn(suite)

			return nil
		}, false},
		{"valid past update before client was frozen", func() error {
			clientState = ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, testClientHeight, ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false)
			clientState.FrozenHeight = types.NewHeight(0, testClientHeight.VersionHeight-1)
			err := suite.keeper.CreateClient(suite.ctx, testClientID, clientState, suite.consensusState)
			suite.Require().NoError(err)

			height1 := types.NewHeight(0, 1)

			// store previous consensus state
			prevConsState := &ibctmtypes.ConsensusState{
				Timestamp:          suite.past,
				NextValidatorsHash: suite.valSetHash,
			}
			suite.keeper.SetClientConsensusState(suite.ctx, testClientID, height1, prevConsState)

			// updateHeader will fill in consensus state between prevConsState and suite.consState
			// clientState should not be updated
			updateHeader = createPastUpdateFn(suite)
			return nil
		}, true},
		{"invalid header", func() error {
			clientState = ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, testClientHeight, ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false)
			err := suite.keeper.CreateClient(suite.ctx, testClientID, clientState, suite.consensusState)
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

			err = suite.keeper.UpdateClient(suite.ctx, testClientID, updateHeader)

			if tc.expPass {
				suite.Require().NoError(err, err)

				expConsensusState := &ibctmtypes.ConsensusState{
					Timestamp:          updateHeader.GetTime(),
					Root:               commitmenttypes.NewMerkleRoot(updateHeader.Header.GetAppHash()),
					NextValidatorsHash: updateHeader.Header.NextValidatorsHash,
				}

				newClientState, found := suite.keeper.GetClientState(suite.ctx, testClientID)
				suite.Require().True(found, "valid test case %d failed: %s", i, tc.name)

				consensusState, found := suite.keeper.GetClientConsensusState(suite.ctx, testClientID, updateHeader.GetHeight())
				suite.Require().True(found, "valid test case %d failed: %s", i, tc.name)

				// Determine if clientState should be updated or not
				if updateHeader.GetHeight().GT(clientState.GetLatestHeight()) {
					// Header Height is greater than clientState latest Height, clientState should be updated with header.GetHeight()
					suite.Require().Equal(updateHeader.GetHeight(), newClientState.GetLatestHeight(), "clientstate height did not update")
				} else {
					// Update will add past consensus state, clientState should not be updated at all
					suite.Require().Equal(clientState.GetLatestHeight(), newClientState.GetLatestHeight(), "client state height updated for past header")
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
	var localhostClient exported.ClientState = localhosttypes.NewClientState(suite.header.Header.GetChainID(), types.NewHeight(0, uint64(suite.ctx.BlockHeight())))

	suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)

	err := suite.keeper.UpdateClient(suite.ctx, exported.Localhost, nil)
	suite.Require().NoError(err, err)

	clientState, found := suite.keeper.GetClientState(suite.ctx, exported.Localhost)
	suite.Require().True(found)
	suite.Require().Equal(localhostClient.GetLatestHeight().(types.Height).Increment(), clientState.GetLatestHeight())
}

func (suite *KeeperTestSuite) TestUpgradeClient() {
	var (
		upgradedClient exported.ClientState
		upgradeHeight  exported.Height
		clientA        string
		proofUpgrade   []byte
	)

	testCases := []struct {
		name    string
		setup   func()
		expPass bool
	}{
		{
			name: "successful upgrade",
			setup: func() {

				upgradedClient = ibctmtypes.NewClientState("newChainId", ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod+trustingPeriod, maxClockDrift, newClientHeight, ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false)

				// upgrade Height is at next block
				upgradeHeight = clienttypes.NewHeight(0, uint64(suite.chainB.GetContext().BlockHeight()+1))

				// zero custom fields and store in upgrade store
				suite.chainB.App.UpgradeKeeper.SetUpgradedClient(suite.chainB.GetContext(), int64(upgradeHeight.GetVersionHeight()), upgradedClient)

				// commit upgrade store changes and update clients

				suite.coordinator.CommitBlock(suite.chainB)
				err := suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, ibctesting.Tendermint)
				suite.Require().NoError(err)

				cs, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientState(suite.chainA.GetContext(), clientA)
				suite.Require().True(found)

				proofUpgrade, _ = suite.chainB.QueryUpgradeProof(upgradetypes.UpgradedClientKey(int64(upgradeHeight.GetVersionHeight())), cs.GetLatestHeight().GetVersionHeight())
			},
			expPass: true,
		},
		{
			name: "client state not found",
			setup: func() {

				upgradedClient = ibctmtypes.NewClientState("newChainId", ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod+trustingPeriod, maxClockDrift, newClientHeight, ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false)

				// upgrade Height is at next block
				upgradeHeight = clienttypes.NewHeight(0, uint64(suite.chainB.GetContext().BlockHeight()+1))

				// zero custom fields and store in upgrade store
				suite.chainB.App.UpgradeKeeper.SetUpgradedClient(suite.chainB.GetContext(), int64(upgradeHeight.GetVersionHeight()), upgradedClient)

				// commit upgrade store changes and update clients

				suite.coordinator.CommitBlock(suite.chainB)
				err := suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, ibctesting.Tendermint)
				suite.Require().NoError(err)

				cs, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientState(suite.chainA.GetContext(), clientA)
				suite.Require().True(found)

				proofUpgrade, _ = suite.chainB.QueryUpgradeProof(upgradetypes.UpgradedClientKey(int64(upgradeHeight.GetVersionHeight())), cs.GetLatestHeight().GetVersionHeight())

				clientA = "wrongclientid"
			},
			expPass: false,
		},
		{
			name: "client state frozen",
			setup: func() {

				upgradedClient = ibctmtypes.NewClientState("newChainId", ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod+trustingPeriod, maxClockDrift, newClientHeight, ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false)

				// upgrade Height is at next block
				upgradeHeight = clienttypes.NewHeight(0, uint64(suite.chainB.GetContext().BlockHeight()+1))

				// zero custom fields and store in upgrade store
				suite.chainB.App.UpgradeKeeper.SetUpgradedClient(suite.chainB.GetContext(), int64(upgradeHeight.GetVersionHeight()), upgradedClient)

				// commit upgrade store changes and update clients

				suite.coordinator.CommitBlock(suite.chainB)
				err := suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, ibctesting.Tendermint)
				suite.Require().NoError(err)

				cs, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientState(suite.chainA.GetContext(), clientA)
				suite.Require().True(found)

				proofUpgrade, _ = suite.chainB.QueryUpgradeProof(upgradetypes.UpgradedClientKey(int64(upgradeHeight.GetVersionHeight())), cs.GetLatestHeight().GetVersionHeight())

				// set frozen client in store
				tmClient, ok := cs.(*ibctmtypes.ClientState)
				suite.Require().True(ok)
				tmClient.FrozenHeight = types.NewHeight(0, 1)
				suite.chainA.App.IBCKeeper.ClientKeeper.SetClientState(suite.chainA.GetContext(), clientA, tmClient)
			},
			expPass: false,
		},
		{
			name: "tendermint client VerifyUpgrade fails",
			setup: func() {

				upgradedClient = ibctmtypes.NewClientState("newChainId", ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod+trustingPeriod, maxClockDrift, newClientHeight, ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false)

				// upgrade Height is at next block
				upgradeHeight = clienttypes.NewHeight(0, uint64(suite.chainB.GetContext().BlockHeight()+1))

				// zero custom fields and store in upgrade store
				suite.chainB.App.UpgradeKeeper.SetUpgradedClient(suite.chainB.GetContext(), int64(upgradeHeight.GetVersionHeight()), upgradedClient)

				// change upgradedClient client-specified parameters
				upgradedClient = ibctmtypes.NewClientState("wrongchainID", ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, newClientHeight, ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, true, true)

				suite.coordinator.CommitBlock(suite.chainB)
				err := suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, ibctesting.Tendermint)
				suite.Require().NoError(err)

				cs, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientState(suite.chainA.GetContext(), clientA)
				suite.Require().True(found)

				proofUpgrade, _ = suite.chainB.QueryUpgradeProof(upgradetypes.UpgradedClientKey(int64(upgradeHeight.GetVersionHeight())), cs.GetLatestHeight().GetVersionHeight())
			},
			expPass: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		clientA, _ = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)

		tc.setup()

		err := suite.chainA.App.IBCKeeper.ClientKeeper.UpgradeClient(suite.chainA.GetContext(), clientA, upgradedClient, upgradeHeight, proofUpgrade)

		if tc.expPass {
			suite.Require().NoError(err, "verify upgrade failed on valid case: %s", tc.name)
		} else {
			suite.Require().Error(err, "verify upgrade passed on invalid case: %s", tc.name)
		}
	}

}

func (suite *KeeperTestSuite) TestCheckMisbehaviourAndUpdateState() {
	altPrivVal := ibctestingmock.NewPV()
	altPubKey, err := altPrivVal.GetPubKey()
	suite.Require().NoError(err)
	altVal := tmtypes.NewValidator(altPubKey.(cryptotypes.IntoTmPubKey).AsTmPubKey(), 4)

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
	bothSigners := ibctesting.CreateSortedSignerArray(altPrivVal, suite.privVal, altVal, suiteVal)

	altSigners := []tmtypes.PrivValidator{altPrivVal}

	// Create valid Misbehaviour by making a duplicate header that signs over different block time
	altTime := suite.ctx.BlockTime().Add(time.Minute)

	heightPlus3 := types.NewHeight(0, height+3)
	heightPlus5 := types.NewHeight(0, height+5)

	testCases := []struct {
		name         string
		misbehaviour *ibctmtypes.Misbehaviour
		malleate     func() error
		expPass      bool
	}{
		{
			"trusting period misbehavior should pass",
			&ibctmtypes.Misbehaviour{
				Header1:  suite.chainA.CreateTMClientHeader(testChainID, int64(testClientHeight.VersionHeight), testClientHeight, altTime, bothValSet, bothValSet, bothSigners),
				Header2:  suite.chainA.CreateTMClientHeader(testChainID, int64(testClientHeight.VersionHeight), testClientHeight, suite.ctx.BlockTime(), bothValSet, bothValSet, bothSigners),
				ChainId:  testChainID,
				ClientId: testClientID,
			},
			func() error {
				suite.consensusState.NextValidatorsHash = bothValsHash
				clientState := ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, testClientHeight, ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false)
				err := suite.keeper.CreateClient(suite.ctx, testClientID, clientState, suite.consensusState)

				return err
			},
			true,
		},
		{
			"misbehavior at later height should pass",
			&ibctmtypes.Misbehaviour{
				Header1:  suite.chainA.CreateTMClientHeader(testChainID, int64(heightPlus5.VersionHeight), testClientHeight, altTime, bothValSet, valSet, bothSigners),
				Header2:  suite.chainA.CreateTMClientHeader(testChainID, int64(heightPlus5.VersionHeight), testClientHeight, suite.ctx.BlockTime(), bothValSet, valSet, bothSigners),
				ChainId:  testChainID,
				ClientId: testClientID,
			},
			func() error {
				suite.consensusState.NextValidatorsHash = valsHash
				clientState := ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, testClientHeight, ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false)
				err := suite.keeper.CreateClient(suite.ctx, testClientID, clientState, suite.consensusState)

				// store intermediate consensus state to check that trustedHeight does not need to be highest consensus state before header height
				intermediateConsState := &ibctmtypes.ConsensusState{
					Timestamp:          suite.now.Add(time.Minute),
					NextValidatorsHash: suite.valSetHash,
				}
				suite.keeper.SetClientConsensusState(suite.ctx, testClientID, heightPlus3, intermediateConsState)

				clientState.LatestHeight = heightPlus3
				suite.keeper.SetClientState(suite.ctx, testClientID, clientState)

				return err
			},
			true,
		},
		{
			"misbehavior at later height with different trusted heights should pass",
			&ibctmtypes.Misbehaviour{
				Header1:  suite.chainA.CreateTMClientHeader(testChainID, int64(heightPlus5.VersionHeight), testClientHeight, altTime, bothValSet, valSet, bothSigners),
				Header2:  suite.chainA.CreateTMClientHeader(testChainID, int64(heightPlus5.VersionHeight), heightPlus3, suite.ctx.BlockTime(), bothValSet, bothValSet, bothSigners),
				ChainId:  testChainID,
				ClientId: testClientID,
			},
			func() error {
				suite.consensusState.NextValidatorsHash = valsHash
				clientState := ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, testClientHeight, ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false)
				err := suite.keeper.CreateClient(suite.ctx, testClientID, clientState, suite.consensusState)

				// store trusted consensus state for Header2
				intermediateConsState := &ibctmtypes.ConsensusState{
					Timestamp:          suite.now.Add(time.Minute),
					NextValidatorsHash: bothValsHash,
				}
				suite.keeper.SetClientConsensusState(suite.ctx, testClientID, heightPlus3, intermediateConsState)

				clientState.LatestHeight = heightPlus3
				suite.keeper.SetClientState(suite.ctx, testClientID, clientState)

				return err
			},
			true,
		},
		{
			"trusted ConsensusState1 not found",
			&ibctmtypes.Misbehaviour{
				Header1:  suite.chainA.CreateTMClientHeader(testChainID, int64(heightPlus5.VersionHeight), heightPlus3, altTime, bothValSet, bothValSet, bothSigners),
				Header2:  suite.chainA.CreateTMClientHeader(testChainID, int64(heightPlus5.VersionHeight), testClientHeight, suite.ctx.BlockTime(), bothValSet, valSet, bothSigners),
				ChainId:  testChainID,
				ClientId: testClientID,
			},
			func() error {
				suite.consensusState.NextValidatorsHash = valsHash
				clientState := ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, testClientHeight, ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false)
				err := suite.keeper.CreateClient(suite.ctx, testClientID, clientState, suite.consensusState)
				// intermediate consensus state at height + 3 is not created
				return err
			},
			false,
		},
		{
			"trusted ConsensusState2 not found",
			&ibctmtypes.Misbehaviour{
				Header1:  suite.chainA.CreateTMClientHeader(testChainID, int64(heightPlus5.VersionHeight), testClientHeight, altTime, bothValSet, valSet, bothSigners),
				Header2:  suite.chainA.CreateTMClientHeader(testChainID, int64(heightPlus5.VersionHeight), heightPlus3, suite.ctx.BlockTime(), bothValSet, bothValSet, bothSigners),
				ChainId:  testChainID,
				ClientId: testClientID,
			},
			func() error {
				suite.consensusState.NextValidatorsHash = valsHash
				clientState := ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, testClientHeight, ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false)
				err := suite.keeper.CreateClient(suite.ctx, testClientID, clientState, suite.consensusState)
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
				Header1:  suite.chainA.CreateTMClientHeader(testChainID, int64(testClientHeight.VersionHeight), testClientHeight, altTime, bothValSet, bothValSet, bothSigners),
				Header2:  suite.chainA.CreateTMClientHeader(testChainID, int64(testClientHeight.VersionHeight), testClientHeight, suite.ctx.BlockTime(), bothValSet, bothValSet, bothSigners),
				ChainId:  testChainID,
				ClientId: testClientID,
			},
			func() error {
				suite.consensusState.NextValidatorsHash = bothValsHash
				clientState := ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, testClientHeight, ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false)
				err := suite.keeper.CreateClient(suite.ctx, testClientID, clientState, suite.consensusState)

				clientState.FrozenHeight = types.NewHeight(0, 1)
				suite.keeper.SetClientState(suite.ctx, testClientID, clientState)

				return err
			},
			false,
		},
		{
			"misbehaviour check failed",
			&ibctmtypes.Misbehaviour{
				Header1:  suite.chainA.CreateTMClientHeader(testChainID, int64(testClientHeight.VersionHeight), testClientHeight, altTime, bothValSet, bothValSet, bothSigners),
				Header2:  suite.chainA.CreateTMClientHeader(testChainID, int64(testClientHeight.VersionHeight), testClientHeight, suite.ctx.BlockTime(), altValSet, bothValSet, altSigners),
				ChainId:  testChainID,
				ClientId: testClientID,
			},
			func() error {
				clientState := ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, testClientHeight, ibctesting.DefaultConsensusParams, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false)
				if err != nil {
					return err
				}
				err = suite.keeper.CreateClient(suite.ctx, testClientID, clientState, suite.consensusState)

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

			err = suite.keeper.CheckMisbehaviourAndUpdateState(suite.ctx, tc.misbehaviour)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)

				clientState, found := suite.keeper.GetClientState(suite.ctx, testClientID)
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
