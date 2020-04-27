package types

import (
	"testing"

	"github.com/stretchr/testify/require"
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
		// {
		// 	name: "valid genesis",
		// 	genState: NewGenesisState(
		// 		[]ConnectionEnd{
		// 			{exported.INIT, connectionID, clientID, Counterparty{clientID2, connectionID2, commitmenttypes.NewMerklePrefix([]byte("prefix"))}, []string{"1.0.0"}},
		// 		},
		// 		[]ConnectionPaths{
		// 			{clientID, []string{ibctypes.ConnectionPath(connectionID)}},
		// 		},
		// 	),
		// 	expPass: true,
		// },

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
