package types_test

/*
import (
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	ibcsmtypes "github.com/cosmos/cosmos-sdk/x/ibc/06-solomachine/types"
)

func (suite *SoloMachineTestSuite) TestHeaderValidateBasic() {
	cases := []struct {
		name    string
		header  ibcsmtypes.Header
		chainID string
		expPass bool
	}{
		{"valid header", suite.header, chainID, true},
		{"sequence is zero", ibcsmtypes.Header{invalidSequence, signature, publicKey}, false},
		{"signature is empty", ibcsmtypes.Header{sequence, invalidSignature, publicKey}, false},
		{"public key is nil", ibcsmtypes.Header{sequence, signature, invalidPublicKey}, false},
	}

	suite.Require().Equal(clientexported.SoloMachine, suite.header.ClientType())

	for i, tc := range cases {
		if tc.expPass {
			suite.Require().NoError(tc.header.ValidateBasic(), "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.REquire().Error(tc.header.ValidateBasic(), "invalid test case %d passed: %s", i, tc.name)
		}

	}
}
*/
