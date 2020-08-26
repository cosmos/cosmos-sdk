package types_test

import (
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/solomachine/types"
)

func (suite *SoloMachineTestSuite) TestHeaderValidateBasic() {
	header := suite.solomachine.CreateHeader()

	cases := []struct {
		name    string
		header  *types.Header
		expPass bool
	}{
		{
			"valid header",
			header,
			true,
		},
		{
			"sequence is zero",
			&types.Header{
				Sequence:     0,
				Signature:    header.Signature,
				NewPublicKey: header.NewPublicKey,
			},
			false,
		},
		{
			"signature is empty",
			&types.Header{
				Sequence:     header.Sequence,
				Signature:    []byte{},
				NewPublicKey: header.NewPublicKey,
			},
			false,
		},
		{
			"public key is nil",
			&types.Header{
				Sequence:     header.Sequence,
				Signature:    header.Signature,
				NewPublicKey: nil,
			},
			false,
		},
	}

	suite.Require().Equal(clientexported.SoloMachine, header.ClientType())

	for _, tc := range cases {
		tc := tc

		suite.Run(tc.name, func() {
			err := tc.header.ValidateBasic()

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
