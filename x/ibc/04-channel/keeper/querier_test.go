package keeper_test

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

// TestQueryChannels tests singular, muliple, and no connection for
// correct retrieval of all channels.
func (suite *KeeperTestSuite) TestQueryChannels() {
	path := []string{types.SubModuleName, types.QueryAllChannels}
	var (
		expRes []byte
		err    error
	)

	params := types.NewQueryAllChannelsParams(1, 100)
	data, err := suite.cdc.MarshalJSON(params)
	suite.NoError(err)

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
				suite.NoError(err)
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
				suite.NoError(err)
			},
		},
		{
			"success no channels",
			func() {
				suite.SetupTest()
				expRes, err = codec.MarshalJSONIndent(suite.cdc, []types.IdentifiedChannel{})
				suite.NoError(err)
			},
		},
	}

	for i, tc := range testCases {
		tc.setup()

		bz, err := suite.querier(suite.chainA.GetContext(), path, query)

		suite.NoError(err, "test case %d failed: %s", i, tc.name)
		suite.Equal(expRes, bz, "test case %d failed: %s", i, tc.name)
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
	suite.NoError(err)

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
				suite.NoError(err)
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
				suite.NoError(err)
			},
		},
		{
			"success no channels",
			func() {
				suite.SetupTest()
				expRes, err = codec.MarshalJSONIndent(suite.cdc, []types.IdentifiedChannel{})
				suite.NoError(err)
			},
		},
	}

	for i, tc := range testCases {
		tc.setup()

		bz, err := suite.querier(suite.chainA.GetContext(), path, query)

		suite.NoError(err, "test case %d failed: %s", i, tc.name)
		suite.Equal(expRes, bz, "test case %d failed: %s", i, tc.name)
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
	suite.NoError(err)

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
				suite.NoError(err)
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
				suite.NoError(err)
			},
		},
		{
			"success no packet commitments",
			func() {
				suite.SetupTest()
				expRes, err = codec.MarshalJSONIndent(suite.cdc, []uint64{})
				suite.NoError(err)
			},
		},
	}

	for i, tc := range testCases {
		tc.setup()

		bz, err := suite.querier(suite.chainA.GetContext(), path, query)

		suite.NoError(err, "test case %d failed: %s", i, tc.name)
		suite.Equal(expRes, bz, "test case %d failed: %s", i, tc.name)
	}

}

/*
// TestQueryUnrelayedAcks tests querying unrelayed acknowledgements on a specified channel end.
func (suite *KeeperTestSuite) TestQueryUnrelayedAcks() {
	var (
		expRes []byte
	)

	query := abci.RequestQuery{
		Path: []string{channel.SubModuleName, channel.QueryChannel},
		Data: []byte{},
	}

	testCases := []struct {
		name  string
		setup func()
	}{
		{
			"success",
			func() {
				// create channels on different connections
				// add to expected result
			},
		},
		{
			"success with multiple channels",
			func() {
			},
		},
		{
			"success no unrelayed acks",
			func() {

			},
		},
	}

	for i, tc := range cases {
		tc.setup()

		bz, err := suite.querier(suite.chainA.GetContext(), tc.path, query)

		suite.NoError(err, "test case %d failed: %s", i, tc.name)
		suite.Equal(expRes, bz, "test case %d failed: %s", i, tc.name)
	}

}
*/
