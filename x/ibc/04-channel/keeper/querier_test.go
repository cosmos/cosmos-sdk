package keeper_test

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
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
	data, err := suite.cdc.MarshalJSON(params)
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
				suite.SetupTest()
				channels := make([]types.IdentifiedChannel, 0, 2)

				// create channels on different connections
				suite.chainA.createConnection(
					testConnectionIDA, testConnectionIDB,
					testClientIDA, testClientIDB,
					connection.OPEN,
				)
				channels = append(channels,
					types.NewIdentifiedChannel(testPort1, testChannel1,
						suite.chainA.createChannel(testPort1, testChannel1, testPort2, testChannel2,
							types.OPEN, types.ORDERED, testConnectionIDA,
						),
					),
				)

				suite.chainA.createConnection(
					testConnectionIDB, testConnectionIDA,
					testClientIDB, testClientIDA,
					connection.OPEN,
				)
				channels = append(channels,
					types.NewIdentifiedChannel(testPort2, testChannel2,
						suite.chainA.createChannel(testPort2, testChannel2, testPort1, testChannel1,
							types.OPEN, types.ORDERED, testConnectionIDB,
						),
					),
				)

				// set expected result
				expRes, err = codec.MarshalJSONIndent(suite.cdc, channels)
				suite.Require().NoError(err)
			},
		},
		{
			"success with singular connection channels",
			func() {
				suite.SetupTest()
				channels := make([]types.IdentifiedChannel, 0, 2)

				// create channels on singular connections
				suite.chainA.createConnection(
					testConnectionIDA, testConnectionIDB,
					testClientIDA, testClientIDB,
					connection.OPEN,
				)

				channels = append(channels,
					types.NewIdentifiedChannel(testPort1, testChannel1,
						suite.chainA.createChannel(testPort1, testChannel1, testPort2, testChannel2,
							types.OPEN, types.ORDERED, testConnectionIDA,
						),
					),
				)
				channels = append(channels,
					types.NewIdentifiedChannel(testPort2, testChannel2,
						suite.chainA.createChannel(testPort2, testChannel2, testPort1, testChannel1,
							types.OPEN, types.UNORDERED, testConnectionIDA,
						),
					),
				)

				// set expected result
				expRes, err = codec.MarshalJSONIndent(suite.cdc, channels)
				suite.Require().NoError(err)
			},
		},
		{
			"success no channels",
			func() {
				suite.SetupTest()
				expRes, err = codec.MarshalJSONIndent(suite.cdc, []types.IdentifiedChannel{})
				suite.Require().NoError(err)
			},
		},
	}

	for i, tc := range testCases {
		tc.setup()

		bz, err := suite.querier(suite.chainA.GetContext(), path, query)

		suite.Require().NoError(err, "test case %d failed: %s", i, tc.name)
		suite.Require().Equal(expRes, bz, "test case %d failed: %s", i, tc.name)
	}
}

// TestQueryConnectionChannel tests querying existing channels on a singular connection.
func (suite *KeeperTestSuite) TestQueryConnectionChannels() {
	path := []string{types.SubModuleName, types.QueryConnectionChannels}

	var (
		expRes []byte
		err    error
	)

	params := types.NewQueryConnectionChannelsParams(testConnectionIDA, 1, 100)
	data, err := suite.cdc.MarshalJSON(params)
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
			"success with singular connection channels",
			func() {
				suite.SetupTest()
				channels := make([]types.IdentifiedChannel, 0, 2)

				// create channels on singular connections
				suite.chainA.createConnection(
					testConnectionIDA, testConnectionIDB,
					testClientIDA, testClientIDB,
					connection.OPEN,
				)

				channels = append(channels,
					types.NewIdentifiedChannel(testPort1, testChannel1,
						suite.chainA.createChannel(testPort1, testChannel1, testPort2, testChannel2,
							types.OPEN, types.ORDERED, testConnectionIDA,
						),
					),
				)
				channels = append(channels,
					types.NewIdentifiedChannel(testPort2, testChannel2,
						suite.chainA.createChannel(testPort2, testChannel2, testPort1, testChannel1,
							types.OPEN, types.UNORDERED, testConnectionIDA,
						),
					),
				)

				// set expected result
				expRes, err = codec.MarshalJSONIndent(suite.cdc, channels)
				suite.Require().NoError(err)
			},
		},
		{
			"success multiple connection channels",
			func() {
				suite.SetupTest()
				channels := make([]types.IdentifiedChannel, 0, 1)

				// create channels on different connections
				suite.chainA.createConnection(
					testConnectionIDA, testConnectionIDB,
					testClientIDA, testClientIDB,
					connection.OPEN,
				)
				channels = append(channels,
					types.NewIdentifiedChannel(testPort1, testChannel1,
						suite.chainA.createChannel(testPort1, testChannel1, testPort2, testChannel2,
							types.OPEN, types.ORDERED, testConnectionIDA,
						),
					),
				)

				suite.chainA.createConnection(
					testConnectionIDB, testConnectionIDA,
					testClientIDB, testClientIDA,
					connection.OPEN,
				)
				suite.chainA.createChannel(
					testPort2, testChannel2, testPort1, testChannel1,
					types.OPEN, types.ORDERED, testConnectionIDB,
				)

				// set expected result
				expRes, err = codec.MarshalJSONIndent(suite.cdc, channels)
				suite.Require().NoError(err)
			},
		},
		{
			"success no channels",
			func() {
				suite.SetupTest()
				expRes, err = codec.MarshalJSONIndent(suite.cdc, []types.IdentifiedChannel{})
				suite.Require().NoError(err)
			},
		},
	}

	for i, tc := range testCases {
		tc.setup()

		bz, err := suite.querier(suite.chainA.GetContext(), path, query)

		suite.Require().NoError(err, "test case %d failed: %s", i, tc.name)
		suite.Require().Equal(expRes, bz, "test case %d failed: %s", i, tc.name)
	}
}

func (suite *KeeperTestSuite) TestQuerierChannelClientState() {
	path := []string{types.SubModuleName, types.QueryChannelClientState}
	params := types.NewQueryChannelClientStateParams(testPort1, testChannel1)
	data, err := suite.cdc.MarshalJSON(params)
	suite.Require().NoError(err)

	query := abci.RequestQuery{
		Path: "",
		Data: data,
	}

	testCases := []struct {
		name    string
		setup   func()
		expPass bool
	}{
		{
			"channel not found",
			func() {},
			false,
		},
		{
			"connection for channel not found",
			func() {
				_ = suite.chainA.createChannel(
					testPort1, testChannel1, testPort2, testChannel2,
					types.OPEN, types.ORDERED, testConnectionIDA,
				)
			},
			false,
		},
		{
			"client state for channel's connection not found",
			func() {
				_ = suite.chainA.createConnection(
					testConnectionIDA, testConnectionIDB,
					testClientIDA, testClientIDB,
					connection.OPEN,
				)
				_ = suite.chainA.createChannel(
					testPort1, testChannel1, testPort2, testChannel2,
					types.OPEN, types.ORDERED, testConnectionIDA,
				)
			},
			false,
		},
		{
			"success",
			func() {
				err = suite.chainA.CreateClient(suite.chainB)
				suite.Require().NoError(err)
				err = suite.chainB.CreateClient(suite.chainA)
				suite.Require().NoError(err)
				suite.chainA.createConnection(
					testConnectionIDB, testConnectionIDA, testClientIDB, testClientIDA,
					connection.OPEN,
				)
				suite.chainB.createConnection(
					testConnectionIDA, testConnectionIDB, testClientIDA, testClientIDB,
					connection.OPEN,
				)
				suite.chainA.createChannel(
					testPort1, testChannel1, testPort2, testChannel2, types.INIT,
					types.ORDERED, testConnectionIDB,
				)
				suite.chainB.createChannel(
					testPort2, testChannel2, testPort1, testChannel1, types.TRYOPEN,
					types.ORDERED, testConnectionIDA,
				)
			},
			true,
		},
	}

	for i, tc := range testCases {
		tc.setup()

		clientState, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientState(suite.chainA.GetContext(), testClientIDB)
		bz, err := suite.querier(suite.chainA.GetContext(), path, query)

		if tc.expPass {
			// set expected result
			expRes, merr := codec.MarshalJSONIndent(suite.cdc, clientState)
			suite.Require().NoError(merr)
			suite.Require().True(found)
			suite.Require().NoError(err, "test case %d failed: %s", i, tc.name)
			suite.Require().Equal(string(expRes), string(bz), "test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().False(found)
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
	)

	params := types.NewQueryPacketCommitmentsParams(testPort1, testChannel1, 1, 100)
	data, err := suite.cdc.MarshalJSON(params)
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
			"success",
			func() {
				suite.SetupTest()
				ctx := suite.chainA.GetContext()
				seq := uint64(1)
				commitments := []uint64{}

				// create several commitments on the same channel and port
				for i := seq; i < 10; i++ {
					suite.chainA.storePacketCommitment(ctx, testPort1, testChannel1, i)
					commitments = append(commitments, i)
				}

				expRes, err = codec.MarshalJSONIndent(suite.cdc, commitments)
				suite.Require().NoError(err)
			},
		},
		{
			"success with multiple channels",
			func() {
				suite.SetupTest()
				ctx := suite.chainA.GetContext()
				seq := uint64(1)
				commitments := []uint64{}

				// create several commitments on the same channel and port
				for i := seq; i < 10; i++ {
					suite.chainA.storePacketCommitment(ctx, testPort1, testChannel1, i)
					commitments = append(commitments, i)
				}

				// create several commitments on a different channel and port
				for i := seq; i < 10; i++ {
					suite.chainA.storePacketCommitment(ctx, testPort2, testChannel2, i)
				}

				expRes, err = codec.MarshalJSONIndent(suite.cdc, commitments)
				suite.Require().NoError(err)
			},
		},
		{
			"success no packet commitments",
			func() {
				suite.SetupTest()
				expRes, err = codec.MarshalJSONIndent(suite.cdc, []uint64{})
				suite.Require().NoError(err)
			},
		},
	}

	for i, tc := range testCases {
		tc.setup()

		bz, err := suite.querier(suite.chainA.GetContext(), path, query)

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
	)

	params := types.NewQueryUnrelayedPacketsParams(testPort1, testChannel1, sequences, 1, 100)
	data, err := suite.cdc.MarshalJSON(params)
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
			"success",
			func() {
				suite.SetupTest()
				ctx := suite.chainA.GetContext()
				unrelayedAcks := []uint64{}
				unrelayedSends := []uint64{}

				// create acknowledgements for first 3 sequences
				for _, seq := range sequences {
					if seq < 4 {
						suite.chainA.storeAcknowledgement(ctx, testPort1, testChannel1, seq)
						unrelayedAcks = append(unrelayedAcks, seq)
					} else {
						unrelayedSends = append(unrelayedSends, seq)
					}
				}

				expResAck, err = codec.MarshalJSONIndent(suite.cdc, unrelayedAcks)
				suite.Require().NoError(err)

				expResSend, err = codec.MarshalJSONIndent(suite.cdc, unrelayedSends)
				suite.Require().NoError(err)

			},
		},
		{
			"success with multiple channels",
			func() {
				suite.SetupTest()
				ctx := suite.chainA.GetContext()
				unrelayedAcks := []uint64{}
				unrelayedSends := []uint64{}

				// create acknowledgements for first 3 sequences
				for _, seq := range sequences {
					if seq < 4 {
						suite.chainA.storeAcknowledgement(ctx, testPort1, testChannel1, seq)
						unrelayedAcks = append(unrelayedAcks, seq)
					} else {
						unrelayedSends = append(unrelayedSends, seq)
					}
				}

				// create acknowledgements for other sequences on different channel/port
				for _, seq := range sequences {
					if seq >= 4 {
						suite.chainA.storeAcknowledgement(ctx, testPort2, testChannel2, seq)
					}
				}

				expResAck, err = codec.MarshalJSONIndent(suite.cdc, unrelayedAcks)
				suite.Require().NoError(err)

				expResSend, err = codec.MarshalJSONIndent(suite.cdc, unrelayedSends)
				suite.Require().NoError(err)
			},
		},
		{
			"success no unrelayed acks",
			func() {
				suite.SetupTest()
				ctx := suite.chainA.GetContext()

				// create acknowledgements for all sequences
				for _, seq := range sequences {
					suite.chainA.storeAcknowledgement(ctx, testPort1, testChannel1, seq)
				}

				expResSend, err = codec.MarshalJSONIndent(suite.cdc, []uint64{})
				suite.Require().NoError(err)

				expResAck, err = codec.MarshalJSONIndent(suite.cdc, sequences)
				suite.Require().NoError(err)
			},
		},
	}

	for i, tc := range testCases {
		tc.setup()

		bz, err := suite.querier(suite.chainA.GetContext(), pathAck, query)

		suite.Require().NoError(err, "test case %d failed: %s", i, tc.name)
		suite.Require().Equal(expResAck, bz, "test case %d failed: %s", i, tc.name)

		bz, err = suite.querier(suite.chainA.GetContext(), pathSend, query)

		suite.Require().NoError(err, "test case %d failed: %s", i, tc.name)
		suite.Require().Equal(expResSend, bz, "test case %d failed: %s", i, tc.name)

	}

}
