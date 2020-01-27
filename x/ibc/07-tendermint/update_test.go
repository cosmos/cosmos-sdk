package tendermint_test

import (
	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

func (suite *TendermintTestSuite) TestCheckValidity() {
	testCases := []struct {
		name        string
		clientState tendermint.ClientState
		chainID     string
		expPass     bool
	}{
		{
			name:        "successful update",
			clientState: tendermint.NewClientState(chainID, height),
			chainID:     chainID,
			expPass:     true,
		},
		{
			name:        "header basic validation failed",
			clientState: tendermint.NewClientState(chainID, height),
			chainID:     "cosmoshub",
			expPass:     false,
		},
		{
			name:        "header height < latest client height",
			clientState: tendermint.NewClientState(chainID, height+1),
			chainID:     chainID,
			expPass:     false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		expectedConsensus := tendermint.ConsensusState{
			Root:             commitment.NewRoot(suite.header.AppHash),
			ValidatorSetHash: suite.header.ValidatorSet.Hash(),
		}

		clientState, consensusState, err := tendermint.CheckValidityAndUpdateState(tc.clientState, suite.header, tc.chainID)

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
