package types_test

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/06-solomachine/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

func (suite *SoloMachineTestSuite) TestHeaderValidateBasic() {
	// test singlesig and multisig public keys
	for _, solomachine := range []*ibctesting.Solomachine{suite.solomachine, suite.solomachineMulti} {

		header := solomachine.CreateHeader()

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
					Sequence:       0,
					Timestamp:      header.Timestamp,
					Signature:      header.Signature,
					NewPublicKey:   header.NewPublicKey,
					NewDiversifier: header.NewDiversifier,
				},
				false,
			},
			{
				"timestamp is zero",
				&types.Header{
					Sequence:       header.Sequence,
					Timestamp:      0,
					Signature:      header.Signature,
					NewPublicKey:   header.NewPublicKey,
					NewDiversifier: header.NewDiversifier,
				},
				false,
			},
			{
				"signature is empty",
				&types.Header{
					Sequence:       header.Sequence,
					Timestamp:      header.Timestamp,
					Signature:      []byte{},
					NewPublicKey:   header.NewPublicKey,
					NewDiversifier: header.NewDiversifier,
				},
				false,
			},
			{
				"diversifier contains only spaces",
				&types.Header{
					Sequence:       header.Sequence,
					Timestamp:      header.Timestamp,
					Signature:      header.Signature,
					NewPublicKey:   header.NewPublicKey,
					NewDiversifier: " ",
				},
				false,
			},
			{
				"public key is nil",
				&types.Header{
					Sequence:       header.Sequence,
					Timestamp:      header.Timestamp,
					Signature:      header.Signature,
					NewPublicKey:   nil,
					NewDiversifier: header.NewDiversifier,
				},
				false,
			},
		}

		suite.Require().Equal(exported.Solomachine, header.ClientType())

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
}
