package types_test

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

func (suite *TendermintTestSuite) TestVerifyUpgrade() {
	var (
		upgradedClient *types.ClientState
		proofUpgrade   []byte
		err            error
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
				suite.chainB.App.UpgradeKeeper.SetUpgradeClient(suite.chainB.GetContext(), upgradedClient)
				cs := suite.chainA.GetClientState(suite.chainA.ClientIDs[0])
				tendermintClient, _ := cs.(*types.ClientState)
				proofUpgrade, err = suite.chainB.QueryProof([]byte(tendermintClient.UpgradePath))
			},
			expPass: true,
		},
	}

}
