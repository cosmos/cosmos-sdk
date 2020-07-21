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
		{"client - QuerierClients",
			[]string{clienttypes.SubModuleName, clienttypes.QueryAllClients},
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
			"connection - invalid query",
			[]string{connectiontypes.SubModuleName, "foo"},
			true,
			fmt.Sprintf("unknown IBC %s query endpoint", connectiontypes.SubModuleName),
		},
		{
			"channel - QuerierChannelClientState",
			[]string{channeltypes.SubModuleName, channeltypes.QueryChannelClientState},
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
