package client_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	client "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/23-commitment/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func (suite *ClientTestSuite) TestNewClientUpdateProposalHandler() {
	var (
		content govtypes.Content
		err     error
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"valid update client proposal", func() {
				clientA, _ := suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)
				clientState := suite.chainA.GetClientState(clientA)

				tmClientState, ok := clientState.(*ibctmtypes.ClientState)
				suite.Require().True(ok)
				tmClientState.AllowUpdateAfterMisbehaviour = true
				tmClientState.FrozenHeight = tmClientState.LatestHeight
				suite.chainA.App.IBCKeeper.ClientKeeper.SetClientState(suite.chainA.GetContext(), clientA, tmClientState)

				// use next header for chainB to update the client on chainA
				header, err := suite.chainA.ConstructUpdateTMClientHeader(suite.chainB, clientA)
				suite.Require().NoError(err)

				content, err = clienttypes.NewClientUpdateProposal(ibctesting.Title, ibctesting.Description, clientA, header)
				suite.Require().NoError(err)
			}, true,
		},
		{
			"nil proposal", func() {
				content = nil
			}, false,
		},
		{
			"unsupported proposal type", func() {
				content = distributiontypes.NewCommunityPoolSpendProposal(ibctesting.Title, ibctesting.Description, suite.chainA.SenderAccount.GetAddress(), sdk.NewCoins(sdk.NewCoin("communityfunds", sdk.NewInt(10))))
			}, false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest() // reset

			tc.malleate()

			proposalHandler := client.NewClientUpdateProposalHandler(suite.chainA.App.IBCKeeper.ClientKeeper)

			err = proposalHandler(suite.chainA.GetContext(), content)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}

}

func (suite *ClientTestSuite) TestUpgradeClient() {
	var (
		clientA        string
		upgradedClient exported.ClientState
		msg            *clienttypes.MsgUpgradeClient
	)

	upgradeHeight := clienttypes.NewHeight(1, 1)

	cases := []struct {
		name    string
		setup   func()
		expPass bool
	}{
		{
			name: "successful upgrade",
			setup: func() {

				upgradedClient = ibctmtypes.NewClientState("newChainId", ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod+ibctesting.TrustingPeriod, ibctesting.MaxClockDrift, upgradeHeight, commitmenttypes.GetSDKSpecs(), &ibctesting.UpgradePath, false, false)
				// zero custom fields and store in upgrade store
				suite.chainB.App.UpgradeKeeper.SetUpgradedClient(suite.chainB.GetContext(), upgradedClient)

				// commit upgrade store changes and update clients

				suite.coordinator.CommitBlock(suite.chainB)
				err := suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, ibctesting.Tendermint)
				suite.Require().NoError(err)

				cs, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientState(suite.chainA.GetContext(), clientA)
				suite.Require().True(found)

				proofUpgrade, _ := suite.chainB.QueryUpgradeProof(upgradetypes.UpgradedClientKey(), cs.GetLatestHeight().GetEpochHeight())

				msg, err = clienttypes.NewMsgUpgradeClient(clientA, upgradedClient, proofUpgrade, suite.chainA.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
			expPass: true,
		},
		{
			name: "invalid upgrade: msg.ClientState does not contain valid clientstate",
			setup: func() {

				cs, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientState(suite.chainA.GetContext(), clientA)
				suite.Require().True(found)

				proofUpgrade, _ := suite.chainB.QueryUpgradeProof(upgradetypes.UpgradedClientKey(), cs.GetLatestHeight().GetEpochHeight())

				consState := ibctmtypes.NewConsensusState(time.Now(), commitmenttypes.NewMerkleRoot([]byte("app_hash")), []byte("next_vals_hash"))
				consAny, err := clienttypes.PackConsensusState(consState)
				suite.Require().NoError(err)

				msg = &types.MsgUpgradeClient{ClientId: clientA, ClientState: consAny, ProofUpgrade: proofUpgrade, Signer: suite.chainA.SenderAccount.GetAddress().String()}
			},
			expPass: false,
		},
		{
			name: "invalid clientstate",
			setup: func() {

				upgradedClient = ibctmtypes.NewClientState("", ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod+ibctesting.TrustingPeriod, ibctesting.MaxClockDrift, upgradeHeight, commitmenttypes.GetSDKSpecs(), &ibctesting.UpgradePath, false, false)
				// zero custom fields and store in upgrade store
				suite.chainB.App.UpgradeKeeper.SetUpgradedClient(suite.chainB.GetContext(), upgradedClient)

				// commit upgrade store changes and update clients

				suite.coordinator.CommitBlock(suite.chainB)
				err := suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, ibctesting.Tendermint)
				suite.Require().NoError(err)

				cs, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientState(suite.chainA.GetContext(), clientA)
				suite.Require().True(found)

				proofUpgrade, _ := suite.chainB.QueryUpgradeProof(upgradetypes.UpgradedClientKey(), cs.GetLatestHeight().GetEpochHeight())

				msg, err = clienttypes.NewMsgUpgradeClient(clientA, upgradedClient, proofUpgrade, suite.chainA.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
			expPass: false,
		},
		{
			name: "VerifyUpgrade fails",
			setup: func() {

				upgradedClient = ibctmtypes.NewClientState("newChainId", ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod+ibctesting.TrustingPeriod, ibctesting.MaxClockDrift, upgradeHeight, commitmenttypes.GetSDKSpecs(), &ibctesting.UpgradePath, false, false)
				// zero custom fields and store in upgrade store
				suite.chainB.App.UpgradeKeeper.SetUpgradedClient(suite.chainB.GetContext(), upgradedClient)

				// commit upgrade store changes and update clients

				suite.coordinator.CommitBlock(suite.chainB)
				err := suite.coordinator.UpdateClient(suite.chainA, suite.chainB, clientA, ibctesting.Tendermint)
				suite.Require().NoError(err)

				msg, err = clienttypes.NewMsgUpgradeClient(clientA, upgradedClient, nil, suite.chainA.SenderAccount.GetAddress())
				suite.Require().NoError(err)
			},
			expPass: false,
		},
	}

	for _, tc := range cases {
		tc := tc
		clientA, _ = suite.coordinator.SetupClients(suite.chainA, suite.chainB, ibctesting.Tendermint)

		tc.setup()

		_, err := client.HandleMsgUpgradeClient(
			suite.chainA.GetContext(), suite.chainA.App.IBCKeeper.ClientKeeper, msg,
		)

		if tc.expPass {
			suite.Require().NoError(err, "upgrade handler failed on valid case: %s", tc.name)
			newClient, ok := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientState(suite.chainA.GetContext(), clientA)
			suite.Require().True(ok)
			suite.Require().Equal(upgradedClient, newClient)
		} else {
			suite.Require().Error(err, "upgrade handler passed on invalid case: %s", tc.name)
		}
	}
}
