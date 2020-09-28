package types_test

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func (suite *TendermintTestSuite) TestVerifyUpgrade() {
	var (
		upgradedClient *types.ClientState
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
