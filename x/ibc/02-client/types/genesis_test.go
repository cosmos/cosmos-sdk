package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
)

const (
	testClientID2 = "ethbridge"

	testClientHeight = 5

	trustingPeriod time.Duration = time.Hour * 24 * 7 * 2
	ubdPeriod      time.Duration = time.Hour * 24 * 7 * 3
	maxClockDrift  time.Duration = time.Second * 10
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
				[]exported.ClientState{
					ibctmtypes.NewClientState(testClientID2, trustingPeriod, ubdPeriod, maxClockDrift, ibctmtypes.Header{}),
				},
				nil,
			),
			expPass: true,
		},
		{
			name: "invalid client",
			genState: types.NewGenesisState(
				[]exported.ClientState{
					ibctmtypes.NewClientState(testClientID2, trustingPeriod, ubdPeriod, maxClockDrift, ibctmtypes.Header{}),
				},
				nil,
			),
			expPass: false,
		},
		{
			name: "invalid consensus state",
			genState: types.NewGenesisState(
				[]exported.ClientState{
					ibctmtypes.NewClientState(testClientID2, trustingPeriod, ubdPeriod, maxClockDrift, ibctmtypes.Header{}),
				},
				nil,
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
