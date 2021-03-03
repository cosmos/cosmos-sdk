package types_test

import "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"

// tests acknowledgement.ValidateBasic and acknowledgement.GetBytes
func (suite TypesTestSuite) TestAcknowledgement() {
	testCases := []struct {
		name    string
		ack     types.Acknowledgement
		expPass bool
	}{
		{
			"valid successful ack",
			types.NewResultAcknowledgement([]byte("success")),
			true,
		},
		{
			"valid failed ack",
			types.NewErrorAcknowledgement("error"),
			true,
		},
		{
			"empty successful ack",
			types.NewResultAcknowledgement([]byte{}),
			false,
		},
		{
			"empty faied ack",
			types.NewErrorAcknowledgement("  "),
			false,
		},
		{
			"nil response",
			types.Acknowledgement{
				Response: nil,
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()

			err := tc.ack.ValidateBasic()

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}

			// expect all acks to be able to be marshaled
			suite.NotPanics(func() {
				bz := tc.ack.GetBytes()
				suite.Require().NotNil(bz)
			})
		})
	}

}
