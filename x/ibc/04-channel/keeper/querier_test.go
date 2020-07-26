package keeper_test

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/KiraCore/cosmos-sdk/codec"
	clientexported "github.com/KiraCore/cosmos-sdk/x/ibc/02-client/exported"
	connectiontypes "github.com/KiraCore/cosmos-sdk/x/ibc/03-connection/types"
	"github.com/KiraCore/cosmos-sdk/x/ibc/04-channel/types"
)

// TestQuerierChannelClientState verifies correct querying of client state associated
// with a channel end.
func (suite *KeeperTestSuite) TestQuerierChannelClientState() {
	path := []string{types.SubModuleName, types.QueryChannelClientState}

	var (
		clientID string
		req      *types.QueryChannelClientStateRequest
	)

	testCases := []struct {
		name    string
		setup   func()
		expPass bool
	}{
		{
			"channel not found",
			func() {
				clientA, err := suite.coordinator.CreateClient(suite.chainA, suite.chainB, clientexported.Tendermint)
				suite.Require().NoError(err)

				clientID = clientA
				req = types.NewQueryChannelClientStateRequest("doesnotexist", "doesnotexist")
			},
			false,
		},
		{
			"connection for channel not found",
			func() {
				// connection for channel is deleted from state
				clientA, _, _, _, channelA, _ := suite.coordinator.Setup(suite.chainA, suite.chainB)

				channel := suite.chainA.GetChannel(channelA)
				channel.ConnectionHops[0] = "doesnotexist"

				// set connection hops to wrong connection ID
				suite.chainA.App.IBCKeeper.ChannelKeeper.SetChannel(suite.chainA.GetContext(), channelA.PortID, channelA.ID, channel)

				clientID = clientA
				req = types.NewQueryChannelClientStateRequest(channelA.PortID, channelA.ID)

			},
			false,
		},
		{
			"client state for channel's connection not found",
			func() {
				clientA, _, connA, _, channelA, _ := suite.coordinator.Setup(suite.chainA, suite.chainB)

				// setting connection to empty results in wrong clientID used
				suite.chainA.App.IBCKeeper.ConnectionKeeper.SetConnection(suite.chainA.GetContext(), connA.ID, connectiontypes.ConnectionEnd{})

				clientID = clientA
				req = types.NewQueryChannelClientStateRequest(channelA.PortID, channelA.ID)
			},
			false,
		},
		{
			"success",
			func() {
				clientA, _, _, _, channelA, _ := suite.coordinator.Setup(suite.chainA, suite.chainB)

				clientID = clientA
				req = types.NewQueryChannelClientStateRequest(channelA.PortID, channelA.ID)
			},
			true,
		},
	}

	for i, tc := range testCases {
		suite.SetupTest() // reset
		tc.setup()

		data, err := suite.chainA.App.AppCodec().MarshalJSON(req)
		suite.Require().NoError(err)

		query := abci.RequestQuery{
			Path: "",
			Data: data,
		}

		clientState, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientState(suite.chainA.GetContext(), clientID)
		bz, err := suite.chainA.Querier(suite.chainA.GetContext(), path, query)

		if tc.expPass {
			// set expected result
			expRes, merr := codec.MarshalJSONIndent(suite.chainA.App.AppCodec(), clientState)
			suite.Require().NoError(merr)
			suite.Require().True(found, "test case %d failed: %s", i, tc.name)
			suite.Require().NoError(err, "test case %d failed: %s", i, tc.name)
			suite.Require().Equal(string(expRes), string(bz), "test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "test case %d passed: %s", i, tc.name)
			suite.Require().Nil(bz, "test case %d passed: %s", i, tc.name)
		}
	}
}
