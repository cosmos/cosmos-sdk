package types_test

import (
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	solomachinetypes "github.com/cosmos/cosmos-sdk/x/ibc/06-solomachine/types"
)

func (suite *SoloMachineTestSuite) TestHeaderValidateBasic() {
	header := suite.CreateHeader()

	cases := []struct {
		name    string
		header  solomachinetypes.Header
		expPass bool
	}{
		{"valid header", header, true},
		{"sequence is zero", solomachinetypes.Header{zero, header.Signature, header.NewPubKey}, false},
		{"signature is empty", solomachinetypes.Header{header.Sequence, []byte{}, header.NewPubKey}, false},
		{"public key is nil", solomachinetypes.Header{header.Sequence, header.Signature, nil}, false},
	}

	suite.Require().Equal(clientexported.SoloMachine, header.ClientType())

	for i, tc := range cases {
		if tc.expPass {
			suite.Require().NoError(tc.header.ValidateBasic(), "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(tc.header.ValidateBasic(), "invalid test case %d passed: %s", i, tc.name)
		}

	}
}
