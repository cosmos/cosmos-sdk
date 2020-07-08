package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

func TestValidateGenesis(t *testing.T) {

	testCases := []struct {
		name     string
		genState types.GenesisState
		expPass  bool
	}{
		{
			name:     "default",
			genState: types.DefaultGenesisState(),
			expPass:  true,
		},
		{
			name: "valid genesis",
			genState: types.NewGenesisState(
				[]types.ConnectionEnd{
					types.NewConnectionEnd(types.INIT, connectionID, clientID, types.Counterparty{clientID2, connectionID2, commitmenttypes.NewMerklePrefix([]byte("prefix"))}, []string{types.DefaultIBCVersion}),
				},
				[]types.ConnectionPaths{
					{clientID, []string{host.ConnectionPath(connectionID)}},
				},
			),
			expPass: true,
		},
		{
			name: "invalid connection",
			genState: types.NewGenesisState(
				[]types.ConnectionEnd{
					types.NewConnectionEnd(types.INIT, connectionID, "(CLIENTIDONE)", types.Counterparty{clientID, connectionID, commitmenttypes.NewMerklePrefix([]byte("prefix"))}, []string{types.DefaultIBCVersion}),
				},
				[]types.ConnectionPaths{
					{clientID, []string{host.ConnectionPath(connectionID)}},
				},
			),
			expPass: false,
		},
		{
			name: "invalid client id",
			genState: types.NewGenesisState(
				[]types.ConnectionEnd{
					types.NewConnectionEnd(types.INIT, connectionID, clientID, types.Counterparty{clientID2, connectionID2, commitmenttypes.NewMerklePrefix([]byte("prefix"))}, []string{types.DefaultIBCVersion}),
				},
				[]types.ConnectionPaths{
					{"(CLIENTIDONE)", []string{host.ConnectionPath(connectionID)}},
				},
			),
			expPass: false,
		},
		{
			name: "invalid path",
			genState: types.NewGenesisState(
				[]types.ConnectionEnd{
					types.NewConnectionEnd(types.INIT, connectionID, clientID, types.Counterparty{clientID2, connectionID2, commitmenttypes.NewMerklePrefix([]byte("prefix"))}, []string{types.DefaultIBCVersion}),
				},
				[]types.ConnectionPaths{
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
