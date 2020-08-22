package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	localhosttypes "github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

const (
	chainID  = "chainID"
	clientID = "ethbridge"

	height = 10
)

var latestTimestamp = time.Date(2020, 01, 01, 20, 34, 58, 651387237, time.UTC)

func TestValidateGenesis(t *testing.T) {
	privVal := tmtypes.NewMockPV()
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err)

	now := time.Now().UTC()

	val := tmtypes.NewValidator(pubKey, 10)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{val})

	header := ibctmtypes.CreateTestHeader(chainID, height, height-1, now, valSet, valSet, []tmtypes.PrivValidator{privVal})

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
				[]types.IdentifiedClientState{
					types.NewIdentifiedClientState(
						clientID, ibctmtypes.NewClientState(chainID, ibctesting.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, height, latestTimestamp, commitmenttypes.GetSDKSpecs(), false, false),
					),
					types.NewIdentifiedClientState(
						exported.ClientTypeLocalHost, localhosttypes.NewClientState("chainID", 10, latestTimestamp),
					),
				},
				[]types.ClientConsensusStates{
					types.NewClientConsensusStates(
						clientID,
						[]exported.ConsensusState{
							ibctmtypes.NewConsensusState(
								header.GetTime(), commitmenttypes.NewMerkleRoot(header.Header.GetAppHash()), header.GetHeight(), header.Header.NextValidatorsHash,
							),
						},
					),
				},
				true,
			),
			expPass: true,
		},
		{
			name: "invalid clientid",
			genState: types.NewGenesisState(
				[]types.IdentifiedClientState{
					types.NewIdentifiedClientState(
						"/~@$*", ibctmtypes.NewClientState(chainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, height, latestTimestamp, commitmenttypes.GetSDKSpecs(), false, false),
					),
					types.NewIdentifiedClientState(
						exported.ClientTypeLocalHost, localhosttypes.NewClientState("chainID", 10, latestTimestamp),
					),
				},
				[]types.ClientConsensusStates{
					types.NewClientConsensusStates(
						clientID,
						[]exported.ConsensusState{
							ibctmtypes.NewConsensusState(
								header.GetTime(), commitmenttypes.NewMerkleRoot(header.Header.GetAppHash()), header.GetHeight(), header.Header.NextValidatorsHash,
							),
						},
					),
				},
				true,
			),
			expPass: false,
		},
		{
			name: "invalid client",
			genState: types.NewGenesisState(
				[]types.IdentifiedClientState{
					types.NewIdentifiedClientState(
						clientID, ibctmtypes.NewClientState(chainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, height, latestTimestamp, commitmenttypes.GetSDKSpecs(), false, false),
					),
					types.NewIdentifiedClientState(exported.ClientTypeLocalHost, localhosttypes.NewClientState("chaindID", 0, latestTimestamp)),
				},
				nil,
				true,
			),
			expPass: false,
		},
		{
			name: "invalid consensus state",
			genState: types.NewGenesisState(
				[]types.IdentifiedClientState{
					types.NewIdentifiedClientState(
						clientID, ibctmtypes.NewClientState(chainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, height, latestTimestamp, commitmenttypes.GetSDKSpecs(), false, false),
					),
					types.NewIdentifiedClientState(
						exported.ClientTypeLocalHost, localhosttypes.NewClientState("chaindID", 10, latestTimestamp),
					),
				},
				[]types.ClientConsensusStates{
					types.NewClientConsensusStates(
						"(CLIENTID2)",
						[]exported.ConsensusState{
							ibctmtypes.NewConsensusState(
								header.GetTime(), commitmenttypes.NewMerkleRoot(header.Header.GetAppHash()), 0, header.Header.NextValidatorsHash,
							),
						},
					),
				},
				true,
			),
			expPass: false,
		},
		{
			name: "invalid consensus state",
			genState: types.NewGenesisState(
				[]types.IdentifiedClientState{
					types.NewIdentifiedClientState(
						clientID, ibctmtypes.NewClientState(chainID, ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, height, latestTimestamp, commitmenttypes.GetSDKSpecs(), false, false),
					),
					types.NewIdentifiedClientState(
						exported.ClientTypeLocalHost, localhosttypes.NewClientState("chaindID", 10, latestTimestamp),
					),
				},
				[]types.ClientConsensusStates{
					types.NewClientConsensusStates(
						clientID,
						[]exported.ConsensusState{
							ibctmtypes.NewConsensusState(
								header.GetTime(), commitmenttypes.NewMerkleRoot(header.Header.GetAppHash()), 0, header.Header.NextValidatorsHash,
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
