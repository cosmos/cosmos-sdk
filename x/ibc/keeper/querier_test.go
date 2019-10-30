package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	"github.com/cosmos/cosmos-sdk/x/ibc/keeper"

	abci "github.com/tendermint/tendermint/abci/types"
)

// TestNewQuerier tests that the querier paths are correct.
// NOTE: the actuall testing functionallity are located on each ICS querier test.
func TestNewQuerier(t *testing.T) {
	app, ctx := createTestApp(true)
	querier := keeper.NewQuerier(app.IBCKeeper)

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
		t.Run(tc.name, func(t *testing.T) {
			_, err := querier(ctx, tc.path, query)
			if tc.expectsDefaultErr {
				require.Contains(t, err.Error(), tc.errMsg, "test case #%d", i)
			} else {
				require.Error(t, err, "test case #%d", i)
			}
		})
	}
}
