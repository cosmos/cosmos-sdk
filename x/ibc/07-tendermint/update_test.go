package tendermint_test

import (
	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

func (suite *TendermintTestSuite) TestCheckValidity() {
	testCases := []struct {
		name        string
		clientState ibctmtypes.ClientState
		chainID     string
		expPass     bool
	}{
		{
			name:        "successful update",
			clientState: ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, height, suite.clientTime),
			chainID:     chainID,
			expPass:     true,
		},
		{
			name:        "header basic validation failed",
			clientState: ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, height, suite.clientTime),
			chainID:     "cosmoshub",
			expPass:     false,
		},
		{
			name:        "header height < latest client height",
			clientState: ibctmtypes.NewClientState(chainID, trustingPeriod, ubdPeriod, height+1, suite.clientTime),
			chainID:     chainID,
			expPass:     false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		expectedConsensus := ibctmtypes.ConsensusState{
			Timestamp:    suite.headerTime,
			Root:         commitment.NewRoot(suite.header.AppHash),
			ValidatorSet: suite.header.ValidatorSet,
		}

		clientState, consensusState, err := tendermint.CheckValidityAndUpdateState(tc.clientState, suite.header, tc.chainID, suite.now)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
			suite.Require().Equal(suite.header.GetHeight(), clientState.GetLatestHeight(), "valid test case %d failed: %s", i, tc.name)
			suite.Require().Equal(expectedConsensus, consensusState, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
			suite.Require().Nil(clientState, "invalid test case %d passed: %s", i, tc.name)
			suite.Require().Nil(consensusState, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}
