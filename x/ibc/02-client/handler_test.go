package client_test

import (
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	localhosttypes "github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/types"
)

func (suite *ClientTestSuite) TestHandleCreateClientLocalHost() {
	cases := []struct {
		name     string
		clientID string
		msg      exported.MsgCreateClient
		expPass  bool
	}{
		{
			"tendermint client",
			"gaiamainnet",
			suite.chainA.ConstructMsgCreateClient(suite.chainB, "gaiamainnet"),
			true,
		},
		{
			"client already exists",
			exported.ClientTypeLocalHost,
			&localhosttypes.MsgCreateClient{suite.chainA.SenderAccount.GetAddress()},
			false,
		},
	}

	for _, tc := range cases {
		_, err := client.HandleMsgCreateClient(
			suite.chainA.GetContext(),
			suite.chainA.App.IBCKeeper.ClientKeeper,
			tc.msg,
		)

		if tc.expPass {
			suite.Require().NoError(err, "expected test case %s to pass, got error %v", tc.name, err)

			clientState, ok := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientState(suite.chainA.GetContext(), tc.clientID)
			suite.Require().True(ok, "could not retrieve clientState")
			suite.Require().NotNil(clientState, "clientstate is nil")
		} else {
			suite.Require().Error(err, "invalid test case %s passed", tc.name)
		}
	}
}
