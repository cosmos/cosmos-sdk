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
		{
			"valid header",
			header,
			true,
		},
		{
			"sequence is zero",
			solomachinetypes.Header{
				Sequence:  0,
				Signature: header.Signature,
				NewPubKey: header.NewPubKey,
			},
			false,
		},
		{
			"signature is empty",
			solomachinetypes.Header{
				Sequence:  header.Sequence,
				Signature: []byte{},
				NewPubKey: header.NewPubKey,
			},
			false,
		},
		{
			"public key is nil",
			solomachinetypes.Header{
				Sequence:  header.Sequence,
				Signature: header.Signature,
				NewPubKey: nil,
			},
			false,
		},
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
