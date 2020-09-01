package types_test

import (
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	localhosttypes "github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"

	"github.com/stretchr/testify/require"
)

type caseAny struct {
	name    string
	any     *codectypes.Any
	expPass bool
}

func TestPackClientState(t *testing.T) {
	testCases := []struct {
		name        string
		clientState exported.ClientState
		expPass     bool
	}{
		{
			"solo machine client",
			ibctesting.NewSolomachine(t, "solomachine").ClientState(),
			true,
		},
		{
			"tendermint client",
			ibctmtypes.NewClientState(chainID, ibctesting.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clientHeight, commitmenttypes.GetSDKSpecs()),
			true,
		},
		{
			"localhost client",
			localhosttypes.NewClientState(chainID, clientHeight),
			true,
		},
		{
			"nil",
			nil,
			false,
		},
	}

	testCasesAny := []caseAny{}

	for _, tc := range testCases {
		clientAny, err := types.PackClientState(tc.clientState)
		if tc.expPass {
			require.NoError(t, err, tc.name)
		} else {
			require.Error(t, err, tc.name)
		}

		testCasesAny = append(testCasesAny, caseAny{tc.name, clientAny, tc.expPass})
	}

	for i, tc := range testCasesAny {
		cs, err := types.UnpackClientState(tc.any)
		if tc.expPass {
			require.NoError(t, err, tc.name)
			require.Equal(t, testCases[i].clientState, cs, tc.name)
		} else {
			require.Error(t, err, tc.name)
		}
	}
}

func TestPackConsensusState(t *testing.T) {
	chain := ibctesting.NewTestChain(t, "cosmoshub")

	testCases := []struct {
		name           string
		consensusState exported.ConsensusState
		expPass        bool
	}{
		{
			"solo machine consensus",
			ibctesting.NewSolomachine(t, "solomachine").ConsensusState(),
			true,
		},
		{
			"tendermint consensus",
			chain.LastHeader.ConsensusState(),
			true,
		},
		{
			"nil",
			nil,
			false,
		},
	}

	testCasesAny := []caseAny{}

	for _, tc := range testCases {
		clientAny, err := types.PackConsensusState(tc.consensusState)
		if tc.expPass {
			require.NoError(t, err, tc.name)
		} else {
			require.Error(t, err, tc.name)
		}
		testCasesAny = append(testCasesAny, caseAny{tc.name, clientAny, tc.expPass})
	}

	for i, tc := range testCasesAny {
		cs, err := types.UnpackConsensusState(tc.any)
		if tc.expPass {
			require.NoError(t, err, tc.name)
			require.Equal(t, testCases[i].consensusState, cs, tc.name)
		} else {
			require.Error(t, err, tc.name)
		}
	}
}

func TestPackHeader(t *testing.T) {
	chain := ibctesting.NewTestChain(t, "cosmoshub")

	testCases := []struct {
		name    string
		header  exported.Header
		expPass bool
	}{
		{
			"solo machine header",
			ibctesting.NewSolomachine(t, "solomachine").CreateHeader(),
			true,
		},
		{
			"tendermint header",
			chain.LastHeader,
			true,
		},
		{
			"nil",
			nil,
			false,
		},
	}

	testCasesAny := []caseAny{}

	for _, tc := range testCases {
		clientAny, err := types.PackHeader(tc.header)
		if tc.expPass {
			require.NoError(t, err, tc.name)
		} else {
			require.Error(t, err, tc.name)
		}

		testCasesAny = append(testCasesAny, caseAny{tc.name, clientAny, tc.expPass})
	}

	for i, tc := range testCasesAny {
		cs, err := types.UnpackHeader(tc.any)
		if tc.expPass {
			require.NoError(t, err, tc.name)
			require.Equal(t, testCases[i].header, cs, tc.name)
		} else {
			require.Error(t, err, tc.name)
		}
	}
}

func TestPackMisbehaviour(t *testing.T) {
	chain := ibctesting.NewTestChain(t, "cosmoshub")

	testCases := []struct {
		name         string
		misbehaviour exported.Misbehaviour
		expPass      bool
	}{
		{
			"solo machine misbehaviour",
			ibctesting.NewSolomachine(t, "solomachine").CreateMisbehaviour(),
			true,
		},
		{
			"tendermint misbehaviour",
			ibctmtypes.NewMisbehaviour("tendermint", chain.ChainID, chain.LastHeader, chain.LastHeader),
			true,
		},
		{
			"nil",
			nil,
			false,
		},
	}

	testCasesAny := []caseAny{}

	for _, tc := range testCases {
		clientAny, err := types.PackMisbehaviour(tc.misbehaviour)
		if tc.expPass {
			require.NoError(t, err, tc.name)
		} else {
			require.Error(t, err, tc.name)
		}

		testCasesAny = append(testCasesAny, caseAny{tc.name, clientAny, tc.expPass})
	}

	for i, tc := range testCasesAny {
		cs, err := types.UnpackMisbehaviour(tc.any)
		if tc.expPass {
			require.NoError(t, err, tc.name)
			require.Equal(t, testCases[i].misbehaviour, cs, tc.name)
		} else {
			require.Error(t, err, tc.name)
		}
	}
}
