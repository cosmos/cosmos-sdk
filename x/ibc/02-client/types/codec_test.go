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
			ibctmtypes.NewClientState(chainID, ibctesting.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, 10, commitmenttypes.GetSDKSpecs()),
			true,
		},
		{
			"localhost client",
			localhosttypes.NewClientState(chainID, height),
			true,
		},
		{
			"nil",
			nil,
			false,
		},
	}

	testCasesAny := []caseAny{}

	for i := range testCases {
		clientAny, err := types.PackClientState(testCases[i].clientState)
		if testCases[i].expPass {
			require.NoError(t, err, testCases[i].name)
		} else {
			require.Error(t, err, testCases[i].name)
		}

		testCasesAny = append(testCasesAny, caseAny{testCases[i].name, clientAny, testCases[i].expPass})
	}

	for i := range testCasesAny {
		cs, err := types.UnpackClientState(testCasesAny[i].any)
		if testCasesAny[i].expPass {
			require.NoError(t, err, testCasesAny[i].name)
			require.Equal(t, testCases[i].clientState, cs, testCasesAny[i].name)
		} else {
			require.Error(t, err, testCasesAny[i].name)
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

	for i := range testCases {
		clientAny, err := types.PackConsensusState(testCases[i].consensusState)
		if testCases[i].expPass {
			require.NoError(t, err, testCases[i].name)
		} else {
			require.Error(t, err, testCases[i].name)
		}
		testCasesAny = append(testCasesAny, caseAny{testCases[i].name, clientAny, testCases[i].expPass})
	}

	for i := range testCasesAny {
		cs, err := types.UnpackConsensusState(testCasesAny[i].any)
		if testCasesAny[i].expPass {
			require.NoError(t, err, testCasesAny[i].name)
			require.Equal(t, testCases[i].consensusState, cs, testCasesAny[i].name)
		} else {
			require.Error(t, err, testCasesAny[i].name)
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

	for i := range testCases {
		clientAny, err := types.PackHeader(testCases[i].header)
		if testCases[i].expPass {
			require.NoError(t, err, testCases[i].name)
		} else {
			require.Error(t, err, testCases[i].name)
		}

		testCasesAny = append(testCasesAny, caseAny{testCases[i].name, clientAny, testCases[i].expPass})
	}

	for i := range testCasesAny {
		cs, err := types.UnpackHeader(testCasesAny[i].any)
		if testCasesAny[i].expPass {
			require.NoError(t, err, testCasesAny[i].name)
			require.Equal(t, testCases[i].header, cs, testCasesAny[i].name)
		} else {
			require.Error(t, err, testCasesAny[i].name)
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

	for i := range testCases {
		clientAny, err := types.PackMisbehaviour(testCases[i].misbehaviour)
		if testCases[i].expPass {
			require.NoError(t, err, testCases[i].name)
		} else {
			require.Error(t, err, testCases[i].name)
		}

		testCasesAny = append(testCasesAny, caseAny{testCases[i].name, clientAny, testCases[i].expPass})
	}

	for i := range testCasesAny {
		cs, err := types.UnpackMisbehaviour(testCasesAny[i].any)
		if testCasesAny[i].expPass {
			require.NoError(t, err, testCasesAny[i].name)
			require.Equal(t, testCases[i].misbehaviour, cs, testCasesAny[i].name)
		} else {
			require.Error(t, err, testCasesAny[i].name)
		}
	}
}
