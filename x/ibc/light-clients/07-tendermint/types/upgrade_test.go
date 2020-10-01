package types_test

import (
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
	solomachinetypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/06-solomachine/types"
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
			name: "successful upgrade to a new tendermint client",
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
			name: "successful upgrade to a new tendermint client with different client chosen parameters",
			setup: func() {

				upgradedClient = types.NewClientState("newChainId", types.DefaultTrustLevel, trustingPeriod, ubdPeriod+trustingPeriod, maxClockDrift, upgradeHeight, commitmenttypes.GetSDKSpecs(), &upgradePath, false, false)
				// zero custom fields and store in upgrade store
				suite.chainB.App.UpgradeKeeper.SetUpgradedClient(suite.chainB.GetContext(), upgradedClient)

				// change upgradedClient client-specified parameters
				upgradedClient = types.NewClientState("newChainId", types.DefaultTrustLevel, ubdPeriod, ubdPeriod+trustingPeriod, maxClockDrift+5, upgradeHeight, commitmenttypes.GetSDKSpecs(), &upgradePath, true, true)

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
			name: "successful upgrade to a solomachine client",
			setup: func() {
				cs, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientState(suite.chainA.GetContext(), clientA)
				suite.Require().True(found)

				// demonstrate that VerifyUpgrade allows for arbitrary changes to clienstate structure so long as
				// previous chain committed to the change
				upgradedClient = ibctesting.NewSolomachine(suite.T(), suite.cdc, clientA, "diversifier", 1).ClientState()
				soloClient, _ := upgradedClient.(*solomachinetypes.ClientState)
				// change sequence to be higher height than latest current client height
				soloClient.Sequence = cs.GetLatestHeight().GetEpochHeight() + 100
				// zero custom fields and store in upgrade store
				suite.chainB.App.UpgradeKeeper.SetUpgradedClient(suite.chainB.GetContext(), upgradedClient)

				suite.coordinator.CommitBlock(suite.chainB)
				err := suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, ibctesting.Tendermint)
				suite.Require().NoError(err)

				cs, found = suite.chainA.App.IBCKeeper.ClientKeeper.GetClientState(suite.chainA.GetContext(), clientA)
				suite.Require().True(found)

				proofUpgrade, _ = suite.chainB.QueryUpgradeProof(upgradetypes.UpgradedClientKey(), cs.GetLatestHeight().GetEpochHeight())
			},
			expPass: true,
		},
		{
			name: "successful upgrade to a solomachine client with different client-chosen parameters",
			setup: func() {
				cs, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientState(suite.chainA.GetContext(), clientA)
				suite.Require().True(found)

				// demonstrate that VerifyUpgrade allows for arbitrary changes to clienstate structure so long as
				// previous chain committed to the change
				upgradedClient = ibctesting.NewSolomachine(suite.T(), suite.cdc, clientA, "diversifier", 1).ClientState()
				soloClient, _ := upgradedClient.(*solomachinetypes.ClientState)
				// change sequence to be higher height than latest current client height
				soloClient.Sequence = cs.GetLatestHeight().GetEpochHeight() + 100
				// zero custom fields and store in upgrade store
				suite.chainB.App.UpgradeKeeper.SetUpgradedClient(suite.chainB.GetContext(), soloClient)

				// change client-specified parameter
				soloClient.AllowUpdateAfterProposal = true

				suite.coordinator.CommitBlock(suite.chainB)
				err := suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, ibctesting.Tendermint)
				suite.Require().NoError(err)

				cs, found = suite.chainA.App.IBCKeeper.ClientKeeper.GetClientState(suite.chainA.GetContext(), clientA)
				suite.Require().True(found)

				proofUpgrade, _ = suite.chainB.QueryUpgradeProof(upgradetypes.UpgradedClientKey(), cs.GetLatestHeight().GetEpochHeight())
			},
			expPass: true,
		},
		{
			name: "unsuccessful upgrade to a new tendermint client: chain-specified paramaters do not match committed client",
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

			name: "unsuccessful upgrade to a new solomachine client: chain-specified paramaters do not match committed client",
			setup: func() {
				cs, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientState(suite.chainA.GetContext(), clientA)
				suite.Require().True(found)

				// demonstrate that VerifyUpgrade allows for arbitrary changes to clienstate structure so long as
				// previous chain committed to the change
				upgradedClient = ibctesting.NewSolomachine(suite.T(), suite.cdc, clientA, "diversifier", 1).ClientState()
				soloClient, _ := upgradedClient.(*solomachinetypes.ClientState)
				// change sequence to be higher height than latest current client height
				soloClient.Sequence = cs.GetLatestHeight().GetEpochHeight() + 100
				// zero custom fields and store in upgrade store
				suite.chainB.App.UpgradeKeeper.SetUpgradedClient(suite.chainB.GetContext(), upgradedClient)

				// change chain-specified parameters from committed values
				soloClient.Sequence = 10000

				suite.coordinator.CommitBlock(suite.chainB)
				err := suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, ibctesting.Tendermint)
				suite.Require().NoError(err)

				cs, found = suite.chainA.App.IBCKeeper.ClientKeeper.GetClientState(suite.chainA.GetContext(), clientA)
				suite.Require().True(found)

				proofUpgrade, _ = suite.chainB.QueryUpgradeProof(upgradetypes.UpgradedClientKey(), cs.GetLatestHeight().GetEpochHeight())
			},
			expPass: false,
		},
		{
			name: "unsuccessful upgrade to a new tendermint client: proof is empty",
			setup: func() {
				upgradedClient = types.NewClientState("newChainId", types.DefaultTrustLevel, trustingPeriod, ubdPeriod+trustingPeriod, maxClockDrift, upgradeHeight, commitmenttypes.GetSDKSpecs(), &upgradePath, false, false)
				proofUpgrade = []byte{}
			},
			expPass: false,
		},
		{
			name: "unsuccessful upgrade to a new tendermint client: proof unmarshal failed",
			setup: func() {
				upgradedClient = types.NewClientState("newChainId", types.DefaultTrustLevel, trustingPeriod, ubdPeriod+trustingPeriod, maxClockDrift, upgradeHeight, commitmenttypes.GetSDKSpecs(), &upgradePath, false, false)
				proofUpgrade = []byte("proof")
			},
			expPass: false,
		},
		{
			name: "unsuccessful upgrade to a new tendermint client: proof verification failed",
			setup: func() {
				// create but do not store upgraded client
				upgradedClient = types.NewClientState("newChainId", types.DefaultTrustLevel, trustingPeriod, ubdPeriod+trustingPeriod, maxClockDrift, upgradeHeight, commitmenttypes.GetSDKSpecs(), &upgradePath, false, false)

				cs, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientState(suite.chainA.GetContext(), clientA)
				suite.Require().True(found)

				proofUpgrade, _ = suite.chainB.QueryUpgradeProof(upgradetypes.UpgradedClientKey(), cs.GetLatestHeight().GetEpochHeight())
			},
			expPass: false,
		},
		{
			name: "unsuccessful upgrade to a new tendermint client: upgrade path is nil",
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
			name: "unsuccessful upgrade to a new tendermint client: upgraded height is not greater than current height",
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
	}

	for _, tc := range testCases {
		tc := tc
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
