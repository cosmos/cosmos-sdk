package ibc

import (
	"testing"

	"github.com/stretchr/testify/require"

	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

var (
	connectionID  = "connectionidone"
	clientID      = "clientidone"
	connectionID2 = "connectionidtwo"
	clientID2     = "clientidtwo"
)

func TestValidateGenesis(t *testing.T) {

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
			genState: GenesisState{
				ConnectionGenesis: connection.NewGenesisState(
					[]connection.ConnectionEnd{
						connection.NewConnectionEnd(connectionexported.INIT, connectionID, clientID, connection.NewCounterparty(clientID2, connectionID2, commitmenttypes.NewMerklePrefix([]byte("prefix"))), []string{"1.0.0"}),
					},
					[]connection.ConnectionPaths{
						connection.NewConnectionPaths(clientID, []string{ibctypes.ConnectionPath(connectionID)}),
					},
				),
			},
			expPass: true,
		},
		{
			name: "invalid connection genesis",
			genState: GenesisState{
				ConnectionGenesis: connection.NewGenesisState(
					[]connection.ConnectionEnd{
						connection.NewConnectionEnd(connectionexported.INIT, connectionID, "CLIENTIDONE", connection.NewCounterparty(clientID, connectionID2, commitmenttypes.NewMerklePrefix([]byte("prefix"))), []string{"1.0.0"}),
					},
					nil,
				),
			},
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
