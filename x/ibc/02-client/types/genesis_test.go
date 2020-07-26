package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	lite "github.com/tendermint/tendermint/lite2"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/KiraCore/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/KiraCore/cosmos-sdk/x/ibc/02-client/types"
	ibctmtypes "github.com/KiraCore/cosmos-sdk/x/ibc/07-tendermint/types"
	localhosttypes "github.com/KiraCore/cosmos-sdk/x/ibc/09-localhost/types"
	commitmenttypes "github.com/KiraCore/cosmos-sdk/x/ibc/23-commitment/types"
)

const (
	clientID = "ethbridge"

	trustingPeriod time.Duration = time.Hour * 24 * 7 * 2
	ubdPeriod      time.Duration = time.Hour * 24 * 7 * 3
	maxClockDrift  time.Duration = time.Second * 10
)

func TestValidateGenesis(t *testing.T) {
	privVal := tmtypes.NewMockPV()
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err)

	now := time.Now().UTC()

	val := tmtypes.NewValidator(pubKey, 10)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{val})

	header := ibctmtypes.CreateTestHeader("chainID", 10, now, valSet, []tmtypes.PrivValidator{privVal})

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
					ibctmtypes.NewClientState(clientID, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, header, commitmenttypes.GetSDKSpecs()),
					localhosttypes.NewClientState("chaindID", 10),
				},
				[]types.ClientConsensusStates{
					{
						clientID,
						[]exported.ConsensusState{
							ibctmtypes.NewConsensusState(
								header.Time, commitmenttypes.NewMerkleRoot(header.AppHash), header.GetHeight(), header.ValidatorSet,
							),
						},
					},
				},
				true,
			),
			expPass: true,
		},
		{
			name: "invalid client",
			genState: types.NewGenesisState(
				[]exported.ClientState{
					ibctmtypes.NewClientState(clientID, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, header, commitmenttypes.GetSDKSpecs()),
					localhosttypes.NewClientState("chaindID", 0),
				},
				nil,
				true,
			),
			expPass: false,
		},
		{
			name: "invalid consensus state",
			genState: types.NewGenesisState(
				[]exported.ClientState{
					ibctmtypes.NewClientState(clientID, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, header, commitmenttypes.GetSDKSpecs()),
					localhosttypes.NewClientState("chaindID", 10),
				},
				[]types.ClientConsensusStates{
					{
						"CLIENTID2",
						[]exported.ConsensusState{
							ibctmtypes.NewConsensusState(
								header.Time, commitmenttypes.NewMerkleRoot(header.AppHash), 0, header.ValidatorSet,
							),
						},
					},
				},
				true,
			),
			expPass: false,
		},
		{
			name: "invalid consensus state",
			genState: types.NewGenesisState(
				[]exported.ClientState{
					ibctmtypes.NewClientState(clientID, lite.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, header, commitmenttypes.GetSDKSpecs()),
					localhosttypes.NewClientState("chaindID", 10),
				},
				[]types.ClientConsensusStates{
					types.NewClientConsensusStates(
						clientID,
						[]exported.ConsensusState{
							ibctmtypes.NewConsensusState(
								header.Time, commitmenttypes.NewMerkleRoot(header.AppHash), 0, header.ValidatorSet,
							),
						},
					),
				},
				true,
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
