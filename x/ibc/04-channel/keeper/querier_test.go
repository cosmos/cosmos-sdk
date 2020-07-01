package keeper_test

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

// TestQueryChannels tests singular, multiple, and no connection for
// correct retrieval of all channels.
func (suite *KeeperTestSuite) TestQueryChannels() {
	path := []string{types.SubModuleName, types.QueryAllChannels}
	var (
		expRes []byte
		err    error
	)

	params := types.NewQueryAllChannelsParams(1, 100)
	data, err := suite.chainA.App.AppCodec().MarshalJSON(params)
	suite.Require().NoError(err)

	query := abci.RequestQuery{
		Path: "",
		Data: data,
	}

	testCases := []struct {
		name  string
		setup func()
	}{
		{
			"success with different connection channels",
			func() {
				channels := make([]types.IdentifiedChannel, 0, 2)

				// create first connection/channel
				clientA, clientB, _, _, channelA0, _ := suite.coordinator.Setup(suite.chainA, suite.chainB)

				channels = append(channels,
					types.NewIdentifiedChannel(
						channelA0.PortID,
						channelA0.ID,
						suite.chainA.GetChannel(channelA0),
					),
				)

				// create second connection
				connA1, connB1 := suite.coordinator.CreateConnection(suite.chainA, suite.chainB, clientA, clientB)

				// create second channel on second connection
				channelA1, _ := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA1, connB1, types.ORDERED)

				channels = append(channels,
					types.NewIdentifiedChannel(
						channelA1.PortID,
						channelA1.ID,
						suite.chainA.GetChannel(channelA1),
					),
				)

				// set expected result
				expRes, err = codec.MarshalJSONIndent(suite.chainA.App.AppCodec(), channels)
				suite.Require().NoError(err)
			},
		},
		{
			"success with singular connection channels",
			func() {
				channels := make([]types.IdentifiedChannel, 0, 2)

				// create first connection/channel
				_, _, connA, connB, channelA0, _ := suite.coordinator.Setup(suite.chainA, suite.chainB)

				channels = append(channels,
					types.NewIdentifiedChannel(
						channelA0.PortID,
						channelA0.ID,
						suite.chainA.GetChannel(channelA0),
					),
				)

				// create second channel on the same connection
				channelA1, _ := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA, connB, types.UNORDERED)
				channels = append(channels,
					types.NewIdentifiedChannel(
						channelA1.PortID,
						channelA1.ID,
						suite.chainA.GetChannel(channelA1),
					),
				)

				// set expected result
				expRes, err = codec.MarshalJSONIndent(suite.chainA.App.AppCodec(), channels)
				suite.Require().NoError(err)
			},
		},
		{
			"success no channels",
			func() {
				expRes, err = codec.MarshalJSONIndent(suite.chainA.App.AppCodec(), []types.IdentifiedChannel{})
				suite.Require().NoError(err)
			},
		},
	}

	for i, tc := range testCases {
		suite.SetupTest() // reset
		tc.setup()

		bz, err := suite.chainA.Querier(suite.chainA.GetContext(), path, query)

		suite.Require().NoError(err, "test case %d failed: %s", i, tc.name)
		suite.Require().Equal(expRes, bz, "test case %d failed: %s", i, tc.name)
	}
}

// TestQueryConnectionChannel tests querying existing channels on a singular connection.
func (suite *KeeperTestSuite) TestQueryConnectionChannels() {
	path := []string{types.SubModuleName, types.QueryConnectionChannels}

	var (
		expRes []byte
		params types.QueryConnectionChannelsParams
		err    error
	)

	testCases := []struct {
		name  string
		setup func()
	}{
		{
			"success with singular connection channels",
			func() {
				channels := make([]types.IdentifiedChannel, 0, 2)

				// create first connection/channel
				_, _, connA, connB, channelA0, _ := suite.coordinator.Setup(suite.chainA, suite.chainB)

				channels = append(channels,
					types.NewIdentifiedChannel(
						channelA0.PortID,
						channelA0.ID,
						suite.chainA.GetChannel(channelA0),
					),
				)

				// create second channel on the same connection
				channelA1, _ := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA, connB, types.ORDERED)
				channels = append(channels,
					types.NewIdentifiedChannel(
						channelA1.PortID,
						channelA1.ID,
						suite.chainA.GetChannel(channelA1),
					),
				)

				params = types.NewQueryConnectionChannelsParams(connA.ID, 1, 100)

				// set expected result
				expRes, err = codec.MarshalJSONIndent(suite.chainA.App.AppCodec(), channels)
				suite.Require().NoError(err)
			},
		},
		{
			"success multiple connection channels",
			func() {
				channels := make([]types.IdentifiedChannel, 0, 1)

				// create first connection/channel
				clientA, clientB, connA, _, channelA0, _ := suite.coordinator.Setup(suite.chainA, suite.chainB)

				channels = append(channels,
					types.NewIdentifiedChannel(
						channelA0.PortID,
						channelA0.ID,
						suite.chainA.GetChannel(channelA0),
					),
				)

				// create second connection
				connA1, connB1 := suite.coordinator.CreateConnection(suite.chainA, suite.chainB, clientA, clientB)

				// create second channel on second connection
				suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA1, connB1, types.ORDERED)

				params = types.NewQueryConnectionChannelsParams(connA.ID, 1, 100)

				// set expected result
				expRes, err = codec.MarshalJSONIndent(suite.chainA.App.AppCodec(), channels)
				suite.Require().NoError(err)
			},
		},
		{
			"success no channels",
			func() {
				// create connection but no channels
				_, _, connA, _ := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, clientexported.Tendermint)
				params = types.NewQueryConnectionChannelsParams(connA.ID, 1, 100)

				expRes, err = codec.MarshalJSONIndent(suite.chainA.App.AppCodec(), []types.IdentifiedChannel{})
				suite.Require().NoError(err)
			},
		},
	}

	for i, tc := range testCases {
		suite.SetupTest() // reset
		tc.setup()

		data, err := suite.chainA.App.AppCodec().MarshalJSON(params)
		suite.Require().NoError(err)

		query := abci.RequestQuery{
			Path: "",
			Data: data,
		}

		bz, err := suite.chainA.Querier(suite.chainA.GetContext(), path, query)

		suite.Require().NoError(err, "test case %d failed: %s", i, tc.name)
		suite.Require().Equal(expRes, bz, "test case %d failed: %s", i, tc.name)
	}
}

// TestQuerierChannelClientState verifies correct querying of client state associated
// with a channel end.
func (suite *KeeperTestSuite) TestQuerierChannelClientState() {
	path := []string{types.SubModuleName, types.QueryChannelClientState}

	var (
		clientID string
		params   types.QueryChannelClientStateParams
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
				params = types.NewQueryChannelClientStateParams("doesnotexist", "doesnotexist")
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
				params = types.NewQueryChannelClientStateParams(channelA.PortID, channelA.ID)

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
				params = types.NewQueryChannelClientStateParams(channelA.PortID, channelA.ID)
			},
			false,
		},
		{
			"success",
			func() {
				clientA, _, _, _, channelA, _ := suite.coordinator.Setup(suite.chainA, suite.chainB)

				clientID = clientA
				params = types.NewQueryChannelClientStateParams(channelA.PortID, channelA.ID)
			},
			true,
		},
	}

	for i, tc := range testCases {
		suite.SetupTest() // reset
		tc.setup()

		data, err := suite.chainA.App.AppCodec().MarshalJSON(params)
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

// TestQueryPacketCommitments tests querying packet commitments on a specified channel end.
func (suite *KeeperTestSuite) TestQueryPacketCommitments() {
	path := []string{types.SubModuleName, types.QueryPacketCommitments}

	var (
		expRes []byte
		params types.QueryPacketCommitmentsParams
		err    error
	)

	testCases := []struct {
		name  string
		setup func()
	}{
		{
			"success",
			func() {
				_, _, _, _, channelA, _ := suite.coordinator.Setup(suite.chainA, suite.chainB)

				seq := uint64(1)
				commitments := []uint64{}

				// create several commitments on the same channel and port
				for i := seq; i < 10; i++ {
					suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainA.GetContext(), channelA.PortID, channelA.ID, i, []byte("ack"))
					commitments = append(commitments, i)
				}

				params = types.NewQueryPacketCommitmentsParams(channelA.PortID, channelA.ID, 1, 100)

				expRes, err = codec.MarshalJSONIndent(suite.chainA.App.AppCodec(), commitments)
				suite.Require().NoError(err)
			},
		},
		{
			"success with multiple channels",
			func() {
				_, _, connA, connB, channelA0, _ := suite.coordinator.Setup(suite.chainA, suite.chainB)

				seq := uint64(1)
				commitments := []uint64{}

				// create several commitments on the same channel and port
				for i := seq; i < 10; i++ {
					suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainA.GetContext(), channelA0.PortID, channelA0.ID, i, []byte("ack"))
					commitments = append(commitments, i)
				}

				// create second channel
				channelA1, _ := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA, connB, types.ORDERED)

				// create several commitments on a different channel and port
				for i := seq; i < 10; i++ {
					suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainA.GetContext(), channelA1.PortID, channelA1.ID, i, []byte("ack"))
				}

				params = types.NewQueryPacketCommitmentsParams(channelA0.PortID, channelA1.ID, 1, 100)

				expRes, err = codec.MarshalJSONIndent(suite.chainA.App.AppCodec(), commitments)
				suite.Require().NoError(err)
			},
		},
		{
			"success no packet commitments",
			func() {
				_, _, _, _, channelA, _ := suite.coordinator.Setup(suite.chainA, suite.chainB)

				params = types.NewQueryPacketCommitmentsParams(channelA.PortID, channelA.ID, 1, 100)

				expRes, err = codec.MarshalJSONIndent(suite.chainA.App.AppCodec(), []uint64{})
				suite.Require().NoError(err)
			},
		},
	}

	for i, tc := range testCases {
		suite.SetupTest() // reset
		tc.setup()

		data, err := suite.chainA.App.AppCodec().MarshalJSON(params)
		suite.Require().NoError(err)

		query := abci.RequestQuery{
			Path: "",
			Data: data,
		}

		bz, err := suite.chainA.Querier(suite.chainA.GetContext(), path, query)

		suite.Require().NoError(err, "test case %d failed: %s", i, tc.name)
		suite.Require().Equal(expRes, bz, "test case %d failed: %s", i, tc.name)
	}

}

// TestQueryUnrelayedPackets tests querying unrelayed acknowledgements and unrelayed packets sends
// on a specified channel end.
func (suite *KeeperTestSuite) TestQueryUnrelayedAcks() {
	pathAck := []string{types.SubModuleName, types.QueryUnrelayedAcknowledgements}
	pathSend := []string{types.SubModuleName, types.QueryUnrelayedPacketSends}
	sequences := []uint64{1, 2, 3, 4, 5}

	var (
		expResAck  []byte
		expResSend []byte
		params     types.QueryUnrelayedPacketsParams
		err        error
	)

	testCases := []struct {
		name  string
		setup func()
	}{
		{
			"success",
			func() {
				_, _, _, _, channelA, _ := suite.coordinator.Setup(suite.chainA, suite.chainB)

				unrelayedAcks := []uint64{}
				unrelayedSends := []uint64{}

				// create acknowledgements for first 3 sequences
				for _, seq := range sequences {
					if seq < 4 {
						suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketAcknowledgement(suite.chainA.GetContext(), channelA.PortID, channelA.ID, seq, []byte("ack"))
						unrelayedAcks = append(unrelayedAcks, seq)
					} else {
						unrelayedSends = append(unrelayedSends, seq)
					}
				}

				params = types.NewQueryUnrelayedPacketsParams(channelA.PortID, channelA.ID, sequences, 1, 100)

				expResAck, err = codec.MarshalJSONIndent(suite.chainA.App.AppCodec(), unrelayedAcks)
				suite.Require().NoError(err)

				expResSend, err = codec.MarshalJSONIndent(suite.chainA.App.AppCodec(), unrelayedSends)
				suite.Require().NoError(err)

			},
		},
		{
			"success with multiple channels",
			func() {
				_, _, connA, connB, channelA0, _ := suite.coordinator.Setup(suite.chainA, suite.chainB)
				ctxA := suite.chainA.GetContext()

				unrelayedAcks := []uint64{}
				unrelayedSends := []uint64{}

				// create acknowledgements for first 3 sequences
				for _, seq := range sequences {
					if seq < 4 {
						suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketAcknowledgement(ctxA, channelA0.PortID, channelA0.ID, seq, []byte("ack"))
						unrelayedAcks = append(unrelayedAcks, seq)
					} else {
						unrelayedSends = append(unrelayedSends, seq)
					}
				}

				// create second channel
				channelA1, _ := suite.coordinator.CreateChannel(suite.chainA, suite.chainB, connA, connB, types.UNORDERED)

				// create acknowledgements for other sequences on different channel/port
				for _, seq := range sequences {
					if seq >= 4 {
						suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketAcknowledgement(ctxA, channelA1.PortID, channelA1.ID, seq, []byte("ack"))
					}
				}

				params = types.NewQueryUnrelayedPacketsParams(channelA0.PortID, channelA0.ID, sequences, 1, 100)

				expResAck, err = codec.MarshalJSONIndent(suite.chainA.App.AppCodec(), unrelayedAcks)
				suite.Require().NoError(err)

				expResSend, err = codec.MarshalJSONIndent(suite.chainA.App.AppCodec(), unrelayedSends)
				suite.Require().NoError(err)
			},
		},
		{
			"success no unrelayed acks",
			func() {
				_, _, _, _, channelA, _ := suite.coordinator.Setup(suite.chainA, suite.chainB)

				// create acknowledgements for all sequences
				for _, seq := range sequences {
					suite.chainA.App.IBCKeeper.ChannelKeeper.SetPacketAcknowledgement(suite.chainA.GetContext(), channelA.PortID, channelA.ID, seq, []byte("ack"))
				}

				params = types.NewQueryUnrelayedPacketsParams(channelA.PortID, channelA.ID, sequences, 1, 100)

				expResSend, err = codec.MarshalJSONIndent(suite.chainA.App.AppCodec(), []uint64{})
				suite.Require().NoError(err)

				expResAck, err = codec.MarshalJSONIndent(suite.chainA.App.AppCodec(), sequences)
				suite.Require().NoError(err)
			},
		},
	}

	for i, tc := range testCases {
		suite.SetupTest() // reset
		tc.setup()

		data, err := suite.chainA.App.AppCodec().MarshalJSON(params)
		suite.Require().NoError(err)

		query := abci.RequestQuery{
			Path: "",
			Data: data,
		}

		bz, err := suite.chainA.Querier(suite.chainA.GetContext(), pathAck, query)

		suite.Require().NoError(err, "test case %d failed: %s", i, tc.name)
		suite.Require().Equal(expResAck, bz, "test case %d failed: %s", i, tc.name)

		bz, err = suite.chainA.Querier(suite.chainA.GetContext(), pathSend, query)

		suite.Require().NoError(err, "test case %d failed: %s", i, tc.name)
		suite.Require().Equal(expResSend, bz, "test case %d failed: %s", i, tc.name)

	}

}
