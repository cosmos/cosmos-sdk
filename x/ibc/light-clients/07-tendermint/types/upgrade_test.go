package types_test

import (
	"fmt"

	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/23-commitment/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func (suite *TendermintTestSuite) TestVerifyUpgrade() {
	var (
		upgradedClient exported.ClientState
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

				upgradedClient = types.NewClientState("newChainId", types.DefaultTrustLevel, trustingPeriod, ubdPeriod+trustingPeriod, maxClockDrift, upgradeHeight, commitmenttypes.GetSDKSpecs(), &upgradePath, false, false)
				// zero custom fields and store in upgrade store
				suite.chainB.App.UpgradeKeeper.SetUpgradedClient(suite.chainB.GetContext(), upgradedClient)

				// commit upgrade store changes and update clients

				suite.coordinator.CommitBlock(suite.chainB)
				err := suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, ibctesting.Tendermint)
				suite.Require().NoError(err)

				cs, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientState(suite.chainA.GetContext(), clientA)
				suite.Require().True(found)

				proofUpgrade, _ = suite.chainB.QueryUpgradeProof(upgradetypes.UpgradedClientKey(), cs.GetLatestHeight().GetEpochHeight())
			},
			expPass: true,
		},
		{
			name: "successful upgrade with different client chosen parameters",
			setup: func() {

				upgradedClient = types.NewClientState("newChainId", types.DefaultTrustLevel, trustingPeriod, ubdPeriod+trustingPeriod, maxClockDrift, upgradeHeight, commitmenttypes.GetSDKSpecs(), &upgradePath, false, false)
				// zero custom fields and store in upgrade store
				suite.chainB.App.UpgradeKeeper.SetUpgradedClient(suite.chainB.GetContext(), upgradedClient)

				// change upgradedClient client-specified parameters
				upgradedClient = types.NewClientState("newChainId", types.DefaultTrustLevel, ubdPeriod, ubdPeriod+trustingPeriod, maxClockDrift+5, upgradeHeight, commitmenttypes.GetSDKSpecs(), &upgradePath, true, false)

				suite.coordinator.CommitBlock(suite.chainB)
				err := suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, ibctesting.Tendermint)
				suite.Require().NoError(err)

				cs, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientState(suite.chainA.GetContext(), clientA)
				suite.Require().True(found)

				// Previous client-chosen parameters must be the same as upgraded client chosen parameters
				tmClient, _ := cs.(*types.ClientState)
				oldClient := types.NewClientState(tmClient.ChainId, types.DefaultTrustLevel, ubdPeriod, ubdPeriod+trustingPeriod, maxClockDrift+5, tmClient.LatestHeight, commitmenttypes.GetSDKSpecs(), &upgradePath, true, false)
				suite.chainA.App.IBCKeeper.ClientKeeper.SetClientState(suite.chainA.GetContext(), clientA, oldClient)

				proofUpgrade, _ = suite.chainB.QueryUpgradeProof(upgradetypes.UpgradedClientKey(), cs.GetLatestHeight().GetEpochHeight())
			},
			expPass: true,
		},
		{
			name: "unsuccessful upgrade: chain-specified paramaters do not match committed client",
			setup: func() {

				upgradedClient = types.NewClientState("newChainId", types.DefaultTrustLevel, trustingPeriod, ubdPeriod+trustingPeriod, maxClockDrift, upgradeHeight, commitmenttypes.GetSDKSpecs(), &upgradePath, false, false)
				// zero custom fields and store in upgrade store
				suite.chainB.App.UpgradeKeeper.SetUpgradedClient(suite.chainB.GetContext(), upgradedClient)

				// change upgradedClient client-specified parameters
				upgradedClient = types.NewClientState("wrongchainID", types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, upgradeHeight, commitmenttypes.GetSDKSpecs(), &upgradePath, true, true)

				suite.coordinator.CommitBlock(suite.chainB)
				err := suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, ibctesting.Tendermint)
				suite.Require().NoError(err)

				cs, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientState(suite.chainA.GetContext(), clientA)
				suite.Require().True(found)

				proofUpgrade, _ = suite.chainB.QueryUpgradeProof(upgradetypes.UpgradedClientKey(), cs.GetLatestHeight().GetEpochHeight())
			},
			expPass: false,
		},
		{
			name: "unsuccessful upgrade: client-specified parameters do not match previous client",
			setup: func() {

				upgradedClient = types.NewClientState("newChainId", types.DefaultTrustLevel, trustingPeriod, ubdPeriod+trustingPeriod, maxClockDrift, upgradeHeight, commitmenttypes.GetSDKSpecs(), &upgradePath, false, false)
				// zero custom fields and store in upgrade store
				suite.chainB.App.UpgradeKeeper.SetUpgradedClient(suite.chainB.GetContext(), upgradedClient)

				// change upgradedClient client-specified parameters
				upgradedClient = types.NewClientState("newChainId", types.DefaultTrustLevel, ubdPeriod, ubdPeriod+trustingPeriod, maxClockDrift+5, upgradeHeight, commitmenttypes.GetSDKSpecs(), &upgradePath, true, false)

				suite.coordinator.CommitBlock(suite.chainB)
				err := suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, ibctesting.Tendermint)
				suite.Require().NoError(err)

				cs, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientState(suite.chainA.GetContext(), clientA)
				suite.Require().True(found)

				proofUpgrade, _ = suite.chainB.QueryUpgradeProof(upgradetypes.UpgradedClientKey(), cs.GetLatestHeight().GetEpochHeight())
			},
			expPass: false,
		},
		{
			name: "unsuccessful upgrade: proof is empty",
			setup: func() {
				upgradedClient = types.NewClientState("newChainId", types.DefaultTrustLevel, trustingPeriod, ubdPeriod+trustingPeriod, maxClockDrift, upgradeHeight, commitmenttypes.GetSDKSpecs(), &upgradePath, false, false)
				proofUpgrade = []byte{}
			},
			expPass: false,
		},
		{
			name: "unsuccessful upgrade: proof unmarshal failed",
			setup: func() {
				upgradedClient = types.NewClientState("newChainId", types.DefaultTrustLevel, trustingPeriod, ubdPeriod+trustingPeriod, maxClockDrift, upgradeHeight, commitmenttypes.GetSDKSpecs(), &upgradePath, false, false)
				proofUpgrade = []byte("proof")
			},
			expPass: false,
		},
		{
			name: "unsuccessful upgrade: proof verification failed",
			setup: func() {
				// create but do not store upgraded client
				upgradedClient = types.NewClientState("newChainId", types.DefaultTrustLevel, trustingPeriod, ubdPeriod+trustingPeriod, maxClockDrift, upgradeHeight, commitmenttypes.GetSDKSpecs(), &upgradePath, false, false)

				cs, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientState(suite.chainA.GetContext(), clientA)
				suite.Require().True(found)

				proofUpgrade, _ = suite.chainB.QueryUpgradeProof(upgradetypes.UpgradedClientKey(), cs.GetLatestHeight().GetEpochHeight())
				fmt.Printf("%#v\n", proofUpgrade)
			},
			expPass: false,
		},
		{
			name: "unsuccessful upgrade: upgrade path is nil",
			setup: func() {

				upgradedClient = types.NewClientState("newChainId", types.DefaultTrustLevel, trustingPeriod, ubdPeriod+trustingPeriod, maxClockDrift, upgradeHeight, commitmenttypes.GetSDKSpecs(), &upgradePath, false, false)
				// zero custom fields and store in upgrade store
				suite.chainB.App.UpgradeKeeper.SetUpgradedClient(suite.chainB.GetContext(), upgradedClient)

				// commit upgrade store changes and update clients

				suite.coordinator.CommitBlock(suite.chainB)
				err := suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, ibctesting.Tendermint)
				suite.Require().NoError(err)

				cs, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientState(suite.chainA.GetContext(), clientA)
				suite.Require().True(found)

				proofUpgrade, _ = suite.chainB.QueryUpgradeProof(upgradetypes.UpgradedClientKey(), cs.GetLatestHeight().GetEpochHeight())

				// SetClientState with nil upgrade path
				tmClient, _ := cs.(*types.ClientState)
				tmClient.UpgradePath = nil
				suite.chainA.App.IBCKeeper.ClientKeeper.SetClientState(suite.chainA.GetContext(), clientA, tmClient)
			},
			expPass: false,
		},
		{
			name: "unsuccessful upgrade: upgraded height is not greater than current height",
			setup: func() {

				upgradedClient = types.NewClientState("newChainId", types.DefaultTrustLevel, trustingPeriod, ubdPeriod+trustingPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), &upgradePath, false, false)
				// zero custom fields and store in upgrade store
				suite.chainB.App.UpgradeKeeper.SetUpgradedClient(suite.chainB.GetContext(), upgradedClient)

				// commit upgrade store changes and update clients

				suite.coordinator.CommitBlock(suite.chainB)
				err := suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, ibctesting.Tendermint)
				suite.Require().NoError(err)

				cs, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientState(suite.chainA.GetContext(), clientA)
				suite.Require().True(found)

				proofUpgrade, _ = suite.chainB.QueryUpgradeProof(upgradetypes.UpgradedClientKey(), cs.GetLatestHeight().GetEpochHeight())
			},
			expPass: false,
		},
		{
			name: "unsuccessful upgrade: client is expired",
			setup: func() {

				upgradedClient = types.NewClientState("newChainId", types.DefaultTrustLevel, trustingPeriod, ubdPeriod+trustingPeriod, maxClockDrift, upgradeHeight, commitmenttypes.GetSDKSpecs(), &upgradePath, false, false)
				// zero custom fields and store in upgrade store
				suite.chainB.App.UpgradeKeeper.SetUpgradedClient(suite.chainB.GetContext(), upgradedClient)

				// commit upgrade store changes and update clients

				suite.coordinator.CommitBlock(suite.chainB)
				err := suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, ibctesting.Tendermint)
				suite.Require().NoError(err)

				// expire chainB's client
				suite.chainA.ExpireClient(ubdPeriod)

				cs, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientState(suite.chainA.GetContext(), clientA)
				suite.Require().True(found)

				proofUpgrade, _ = suite.chainB.QueryUpgradeProof(upgradetypes.UpgradedClientKey(), cs.GetLatestHeight().GetEpochHeight())
			},
			expPass: false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		// reset suite
		suite.SetupTest()

		clientA, _ = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)

		tc.setup()

		cs := suite.chainA.GetClientState(clientA)
		clientStore := suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(suite.chainA.GetContext(), clientA)

		err := cs.VerifyUpgrade(
			suite.chainA.GetContext(),
			suite.cdc,
			clientStore,
			upgradedClient,
			proofUpgrade,
		)

		if tc.expPass {
			suite.Require().NoError(err, "verify upgrade failed on valid case: %s", tc.name)
		} else {
			suite.Require().Error(err, "verify upgrade passed on invalid case: %s", tc.name)
		}
	}
}
