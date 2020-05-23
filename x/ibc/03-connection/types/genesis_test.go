package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

func TestValidateGenesis(t *testing.T) {
	prefixAny, err := commitmenttypes.NewMerklePrefix([]byte("prefix")).PackAny()
	require.NoError(t, err)

	testCases := []struct {
		name     string
		genState GenesisState
		expPass  bool
	}{
		{
			name:     "default",
			genState: DefaultGenesisState(),
			expPass:  true,
		},
		{
			name: "valid genesis",
			genState: NewGenesisState(
				[]ConnectionEnd{
					NewConnectionEnd(INIT, connectionID, clientID, Counterparty{clientID2, connectionID2, *prefixAny}, []string{"1.0.0"}),
				},
				[]ConnectionPaths{
					{clientID, []string{host.ConnectionPath(connectionID)}},
				},
			),
			expPass: true,
		},
		{
			name: "invalid connection",
			genState: NewGenesisState(
				[]ConnectionEnd{
					NewConnectionEnd(INIT, connectionID, "(CLIENTIDONE)", Counterparty{clientID, connectionID, *prefixAny}, []string{"1.0.0"}),
				},
				[]ConnectionPaths{
					{clientID, []string{host.ConnectionPath(connectionID)}},
				},
			),
			expPass: false,
		},
		{
			name: "invalid client id",
			genState: NewGenesisState(
				[]ConnectionEnd{
					NewConnectionEnd(INIT, connectionID, clientID, Counterparty{clientID2, connectionID2, *prefixAny}, []string{"1.0.0"}),
				},
				[]ConnectionPaths{
					{"(CLIENTIDONE)", []string{host.ConnectionPath(connectionID)}},
				},
			),
			expPass: false,
		},
		{
			name: "invalid path",
			genState: NewGenesisState(
				[]ConnectionEnd{
					NewConnectionEnd(INIT, connectionID, clientID, Counterparty{clientID2, connectionID2, *prefixAny}, []string{"1.0.0"}),
				},
				[]ConnectionPaths{
					{clientID, []string{connectionID}},
				},
			),
			expPass: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		err := tc.genState.Validate()
		if tc.expPass {
			require.NoError(t, err, tc.name)
		} else {
			require.Error(t, err, tc.name)
		}
	}
}
