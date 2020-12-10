package types_test

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/23-commitment/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
	localhosttypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/09-localhost/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

type caseAny struct {
	name    string
	any     *codectypes.Any
	expPass bool
}

func (suite *TypesTestSuite) TestPackClientState() {

	testCases := []struct {
		name        string
		clientState exported.ClientState
		expPass     bool
	}{
		{
			"solo machine client",
			ibctesting.NewSolomachine(suite.T(), suite.chainA.Codec, "solomachine", "", 2).ClientState(),
			true,
		},
		{
			"tendermint client",
			ibctmtypes.NewClientState(chainID, ibctesting.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clientHeight, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false),
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
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}

		testCasesAny = append(testCasesAny, caseAny{tc.name, clientAny, tc.expPass})
	}

	for i, tc := range testCasesAny {
		cs, err := types.UnpackClientState(tc.any)
		if tc.expPass {
			suite.Require().NoError(err, tc.name)
			suite.Require().Equal(testCases[i].clientState, cs, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}

func (suite *TypesTestSuite) TestPackConsensusState() {
	testCases := []struct {
		name           string
		consensusState exported.ConsensusState
		expPass        bool
	}{
		{
			"solo machine consensus",
			ibctesting.NewSolomachine(suite.T(), suite.chainA.Codec, "solomachine", "", 2).ConsensusState(),
			true,
		},
		{
			"tendermint consensus",
			suite.chainA.LastHeader.ConsensusState(),
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
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
		testCasesAny = append(testCasesAny, caseAny{tc.name, clientAny, tc.expPass})
	}

	for i, tc := range testCasesAny {
		cs, err := types.UnpackConsensusState(tc.any)
		if tc.expPass {
			suite.Require().NoError(err, tc.name)
			suite.Require().Equal(testCases[i].consensusState, cs, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}

func (suite *TypesTestSuite) TestPackHeader() {
	testCases := []struct {
		name    string
		header  exported.Header
		expPass bool
	}{
		{
			"solo machine header",
			ibctesting.NewSolomachine(suite.T(), suite.chainA.Codec, "solomachine", "", 2).CreateHeader(),
			true,
		},
		{
			"tendermint header",
			suite.chainA.LastHeader,
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
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}

		testCasesAny = append(testCasesAny, caseAny{tc.name, clientAny, tc.expPass})
	}

	for i, tc := range testCasesAny {
		cs, err := types.UnpackHeader(tc.any)
		if tc.expPass {
			suite.Require().NoError(err, tc.name)
			suite.Require().Equal(testCases[i].header, cs, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}

func (suite *TypesTestSuite) TestPackMisbehaviour() {
	testCases := []struct {
		name         string
		misbehaviour exported.Misbehaviour
		expPass      bool
	}{
		{
			"solo machine misbehaviour",
			ibctesting.NewSolomachine(suite.T(), suite.chainA.Codec, "solomachine", "", 2).CreateMisbehaviour(),
			true,
		},
		{
			"tendermint misbehaviour",
			ibctmtypes.NewMisbehaviour("tendermint", suite.chainA.LastHeader, suite.chainA.LastHeader),
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
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}

		testCasesAny = append(testCasesAny, caseAny{tc.name, clientAny, tc.expPass})
	}

	for i, tc := range testCasesAny {
		cs, err := types.UnpackMisbehaviour(tc.any)
		if tc.expPass {
			suite.Require().NoError(err, tc.name)
			suite.Require().Equal(testCases[i].misbehaviour, cs, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}
