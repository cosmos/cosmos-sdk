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
// NOTE: the actuall testing functionallity are located on each ICS querier test.
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
		{"client - QueryClientState",
			[]string{client.SubModuleName, client.QueryClientState},
			false,
			"",
		},
		{
			"client - QueryConsensusState",
			[]string{client.SubModuleName, client.QueryConsensusState},
			false,
			"",
		},
		{
			"client - QueryVerifiedRoot",
			[]string{client.SubModuleName, client.QueryVerifiedRoot},
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
			"connection - QueryConnection",
			[]string{connection.SubModuleName, connection.QueryConnection},
			false,
			"",
		},
		{
			"connection - QueryClientConnections",
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
			"channel - QueryChannel",
			[]string{channel.SubModuleName, channel.QueryChannel},
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
