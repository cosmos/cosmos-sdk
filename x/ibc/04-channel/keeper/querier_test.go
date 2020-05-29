package keeper_test

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

// TestQuerier tests all the channel queriers.
func (suite *KeeperTestSuite) TestQueryChannels() {
	path := []string{channel.SubModuleName, channel.QueryAllChannels}
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
				channels := make([]channel.IdentifiedChannel, 0, 2)

				// create channels on different connections
				suite.chainA.createConnection(
					testConnectionIDA, testConnectionIDB,
					testClientIDA, testClientIDB,
					connection.OPEN,
				)
				channels = append(channels,
					types.NewIdentifiedChannel(testPort1, testChannel1,
						suite.chainA.createChannel(testPort1, testChannel1, testPort2, testChannel2,
							channel.OPEN, channel.ORDERED, testConnectionIDA,
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
						suite.chainA.createChannel(testPort2, testChannel2, testPort1, testPort1,
							channel.OPEN, channel.ORDERED, testConnectionIDB,
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
				channels := make([]channel.IdentifiedChannel, 0, 2)

				// create channels on singular connections
				suite.chainA.createConnection(
					testConnectionIDA, testConnectionIDB,
					testClientIDA, testClientIDB,
					connection.OPEN,
				)

				channels = append(channels,
					types.NewIdentifiedChannel(testPort1, testChannel1,
						suite.chainA.createChannel(testPort1, testChannel1, testPort2, testChannel2,
							channel.OPEN, channel.ORDERED, testConnectionIDA,
						),
					),
				)
				channels = append(channels,
					types.NewIdentifiedChannel(testPort2, testChannel2,
						suite.chainA.createChannel(testPort2, testChannel2, testPort1, testChannel1,
							channel.OPEN, channel.UNORDERED, testConnectionIDA,
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
				expRes, err = codec.MarshalJSONIndent(suite.cdc, []channel.IdentifiedChannel{})
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
// TestQueryConnectionChannel tests querying existing channels on a singular connection.
func (suite *KeeperTestSuite) TestQueryConnectionChannels() {
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
			"success with singular connection channels",
			func() {
				// create channels on different connections
				// add to expected result
			},
		},
		{
			"success multiple connection channels",
			func() {
			},
		},
		{
			"success wrong connection, no channels",
			func() {
				expRes = []byte{}
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

// TestQueryPacketCommitments tests querying packet commitments on a specified channel end.
func (suite *KeeperTestSuite) TestQueryPacketCommitments() {
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
			"success no packet commitments",
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
