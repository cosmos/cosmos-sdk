package keeper_test

import (
	"fmt"

	"github.com/stretchr/testify/require"

	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"

	abci "github.com/tendermint/tendermint/abci/types"
)

// TestNewQuerier tests that the querier paths are correct.
// NOTE: the actuall testing functionality are located on each ICS querier test.
func (suite *KeeperTestSuite) TestNewQuerier() {

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	cases := []struct {
		name              string
		path              []string
		expectsDefaultErr bool
		errMsg            string
	}{
		{"client - QuerierClientState",
			[]string{client.SubModuleName, client.QueryClientState},
			false,
			"",
		},
		{"client - QuerierClients",
			[]string{client.SubModuleName, client.QueryAllClients},
			false,
			"",
		},
		{
			"client - QuerierConsensusState",
			[]string{client.SubModuleName, client.QueryConsensusState},
			false,
			"",
		},
		{
			"client - invalid query",
			[]string{client.SubModuleName, "foo"},
			true,
			fmt.Sprintf("unknown IBC %s query endpoint", client.SubModuleName),
		},
		{
			"connection - QuerierConnections",
			[]string{connection.SubModuleName, connection.QueryAllConnections},
			false,
			"",
		},
		{
			"connection - QuerierAllClientConnections",
			[]string{connection.SubModuleName, connection.QueryAllClientConnections},
			false,
			"",
		},
		{
			"connection - QuerierClientConnections",
			[]string{connection.SubModuleName, connection.QueryClientConnections},
			false,
			"",
		},
		{
			"connection - invalid query",
			[]string{connection.SubModuleName, "foo"},
			true,
			fmt.Sprintf("unknown IBC %s query endpoint", connection.SubModuleName),
		},
		{
			"channel - QuerierChannel",
			[]string{channel.SubModuleName, channel.QueryChannel},
			false,
			"",
		},
		{
			"channel - QuerierChannels",
			[]string{channel.SubModuleName, channel.QueryAllChannels},
			false,
			"",
		},
		{
			"channel - QuerierConnectionChannels",
			[]string{channel.SubModuleName, channel.QueryConnectionChannels},
			false,
			"",
		},
		{
			"channel - QuerierPacketCommitments",
			[]string{channel.SubModuleName, channel.QueryPacketCommitments},
			false,
			"",
		},
		{
			"channel - QuerierUnrelayedAcknowledgements",
			[]string{channel.SubModuleName, channel.QueryUnrelayedAcknowledgements},
			false,
			"",
		},
		{
			"channel - QuerierUnrelayedPacketSends",
			[]string{channel.SubModuleName, channel.QueryUnrelayedPacketSends},
			false,
			"",
		},
		{
			"channel - invalid query",
			[]string{channel.SubModuleName, "foo"},
			true,
			fmt.Sprintf("unknown IBC %s query endpoint", channel.SubModuleName),
		},
		{
			"invalid query",
			[]string{"foo"},
			true,
			"unknown IBC query endpoint",
		},
	}

	for i, tc := range cases {
		i, tc := i, tc
		suite.Run(tc.name, func() {
			_, err := suite.querier(suite.ctx, tc.path, query)
			if tc.expectsDefaultErr {
				require.Contains(suite.T(), err.Error(), tc.errMsg, "test case #%d", i)
			} else {
				suite.Error(err, "test case #%d", i)
			}
		})
	}
}
