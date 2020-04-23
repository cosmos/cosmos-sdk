package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateGenesis(t *testing.T) {

	testCases := []struct {
		name     string
		malleate func(*GenesisState)
		expPass  bool
	}{
		{
			name:     "default",
			malleate: func(_ *GenesisState) {},
			expPass:  true,
		},
		{
			name: "valid genesis state",
			malleate: func(genState *GenesisState) {
				genState.Index = 10
				genState.Owners["1"] = CapabilityOwners{[]Owner{{Module: "ibc", Name: "port/transfer"}}}
			},
			expPass: true,
		},
		{
			name: "initial index is 0",
			malleate: func(genState *GenesisState) {
				genState.Index = 0
				genState.Owners["1"] = CapabilityOwners{[]Owner{{Module: "ibc", Name: "port/transfer"}}}
			},
			expPass: false,
		},

		{
			name: "blank owner module",
			malleate: func(genState *GenesisState) {
				genState.Index = 0
				genState.Owners["1"] = CapabilityOwners{[]Owner{{Module: "module", Name: "port/transfer"}}}
			},
			expPass: false,
		},
		{
			name: "blank owner name",
			malleate: func(genState *GenesisState) {
				genState.Index = 0
				genState.Owners["1"] = CapabilityOwners{[]Owner{{Module: "module", Name: ""}}}
			},
			expPass: false,
		},
		{
			name: "index key is not a number",
			malleate: func(genState *GenesisState) {
				genState.Index = 10
				genState.Owners["one"] = CapabilityOwners{[]Owner{{Module: "module", Name: ""}}}
			},
			expPass: false,
		},
		{
			name: "index above range",
			malleate: func(genState *GenesisState) {
				genState.Index = 10
				genState.Owners["12"] = CapabilityOwners{[]Owner{{Module: "module", Name: ""}}}
			},
			expPass: false,
		},
		{
			name: "index below range",
			malleate: func(genState *GenesisState) {
				genState.Index = 10
				genState.Owners["-3"] = CapabilityOwners{[]Owner{{Module: "module", Name: ""}}}
			},
			expPass: false,
		},
		{
			name: "owners are empty",
			malleate: func(genState *GenesisState) {
				genState.Index = 10
				genState.Owners["3"] = CapabilityOwners{[]Owner{}}
			},
			expPass: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		genState := DefaultGenesis()
		tc.malleate(&genState)
		err := ValidateGenesis(genState)
		if tc.expPass {
			require.NoError(t, err, tc.name)
		} else {
			require.Error(t, err, tc.name)
		}
	}
}
