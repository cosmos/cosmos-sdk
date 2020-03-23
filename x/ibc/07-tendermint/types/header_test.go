package types_test

import (
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
)

func (suite *TendermintTestSuite) TestHeaderValidateBasic() {
	testCases := []struct {
		name    string
		header  ibctmtypes.Header
		chainID string
		expPass bool
	}{
		{"valid header", suite.header, chainID, true},
		{"signed header basic validation failed", suite.header, "chainID", false},
		{"validator set nil", ibctmtypes.Header{suite.header.SignedHeader, nil}, chainID, false},
	}

	suite.Require().Equal(clientexported.Tendermint, suite.header.ClientType())

	for i, tc := range testCases {
		tc := tc
		if tc.expPass {
			suite.Require().NoError(tc.header.ValidateBasic(tc.chainID), "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(tc.header.ValidateBasic(tc.chainID), "invalid test case %d passed: %s", i, tc.name)
		}
	}
}
