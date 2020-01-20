package keeper_test

import (
	"fmt"

	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

func (suite *KeeperTestSuite) TestVerifyClientConsensusState() {
	counterparty := types.Counterparty{Prefix: suite.app.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix()}
	connection1 := types.ConnectionEnd{ClientID: testClientID1, Counterparty: counterparty}

	connectionKey := ibctypes.KeyConsensusState(testClientID1, testHeight)

	cases := []struct {
		msg        string
		connection types.ConnectionEnd
		malleate   func() error
		expPass    bool
	}{
		// {"client state not found", connection1, func() error { return nil }, false},
		// {"verification failed", connection1, func() error {
		// 	_, err := suite.app.IBCKeeper.ClientKeeper.CreateClient(suite.ctx, testClientID2, clientexported.Tendermint, suite.consensusState)
		// 	return err
		// }, false},
		{"verification success", connection1, func() error {
			_, err := suite.app.IBCKeeper.ClientKeeper.CreateClient(suite.ctx, testClientID1, clientexported.Tendermint, suite.consensusState)
			return err
		}, true},
	}

	for i, tc := range cases {
		tc := tc
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			err := tc.malleate()
			suite.Require().NoError(err)

			proof, proofHeight := suite.queryProof(connectionKey)

			err = suite.app.IBCKeeper.ConnectionKeeper.VerifyClientConsensusState(
				suite.ctx, tc.connection, uint64(proofHeight), proof, suite.consensusState,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
			}
		})
	}
}
