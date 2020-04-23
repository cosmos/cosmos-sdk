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
			genState: DefaultGenesis(),
			expPass:  true,
		},
		{
			name: "valid genesis state",
			genState: GenesisState{
				Index:  10,
				Owners: []Owner{{Module: "ibc", Name: "port/transfer"}},
			},
			expPass: true,
		},
		{
			name: "invalid index",
			genState: GenesisState{
				Index:  0,
				Owners: []Owner{{Module: "ibc", Name: "port/transfer"}},
			},
			expPass: false,
		},
		{
			name: "blank owner module",
			genState: GenesisState{
				Index:  0,
				Owners: []Owner{{Module: "", Name: "port/transfer"}},
			},
			expPass: false,
		},
		{
			name: "blank owner name",
			genState: GenesisState{
				Index:  10,
				Owners: []Owner{{Module: "ibc", Name: "    "}},
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
