package keeper_test

import (
	"fmt"

	"github.com/stretchr/testify/require"

	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"

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
			[]string{clienttypes.SubModuleName, clienttypes.QueryClientState},
			false,
			"",
		},
		{"client - QuerierClients",
			[]string{clienttypes.SubModuleName, clienttypes.QueryAllClients},
			false,
			"",
		},
		{
			"client - QuerierConsensusState",
			[]string{clienttypes.SubModuleName, clienttypes.QueryConsensusState},
			false,
			"",
		},
		{
			"client - invalid query",
			[]string{clienttypes.SubModuleName, "foo"},
			true,
			fmt.Sprintf("unknown IBC %s query endpoint", clienttypes.SubModuleName),
		},
		{
			"connection - QuerierConnections",
			[]string{connectiontypes.SubModuleName, connectiontypes.QueryAllConnections},
			false,
			"",
		},
		{
			"connection - QuerierAllClientConnections",
			[]string{connectiontypes.SubModuleName, connectiontypes.QueryAllClientConnections},
			false,
			"",
		},
		{
			"connection - QuerierClientConnections",
			[]string{connectiontypes.SubModuleName, connectiontypes.QueryClientConnections},
			false,
			"",
		},
		{
			"connection - invalid query",
			[]string{connectiontypes.SubModuleName, "foo"},
			true,
			fmt.Sprintf("unknown IBC %s query endpoint", connectiontypes.SubModuleName),
		},
		{
			"channel - QuerierChannel",
			[]string{channeltypes.SubModuleName, channeltypes.QueryChannel},
			false,
			"",
		},
		{
			"channel - QuerierChannels",
			[]string{channeltypes.SubModuleName, channeltypes.QueryAllChannels},
			false,
			"",
		},
		{
			"channel - QuerierConnectionChannels",
			[]string{channeltypes.SubModuleName, channeltypes.QueryConnectionChannels},
			false,
			"",
		},
		{
			"channel - QuerierChannelClientState",
			[]string{channeltypes.SubModuleName, channeltypes.QueryChannelClientState},
			false,
			"",
		},
		{
			"channel - QuerierPacketCommitments",
			[]string{channeltypes.SubModuleName, channeltypes.QueryPacketCommitments},
			false,
			"",
		},
		{
			"channel - QuerierUnrelayedAcknowledgements",
			[]string{channeltypes.SubModuleName, channeltypes.QueryUnrelayedAcknowledgements},
			false,
			"",
		},
		{
			"channel - QuerierUnrelayedPacketSends",
			[]string{channeltypes.SubModuleName, channeltypes.QueryUnrelayedPacketSends},
			false,
			"",
		},
		{
			"channel - invalid query",
			[]string{channeltypes.SubModuleName, "foo"},
			true,
			fmt.Sprintf("unknown IBC %s query endpoint", channeltypes.SubModuleName),
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
