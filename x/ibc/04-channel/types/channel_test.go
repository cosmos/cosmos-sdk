package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

func TestCounterpartyValidateBasic(t *testing.T) {
	testCases := []struct {
		name         string
		counterparty types.Counterparty
		expPass      bool
	}{
		{"valid counterparty", types.Counterparty{"portidone", "channelidone"}, true},
		{"invalid port id", types.Counterparty{"(InvalidPort)", "channelidone"}, false},
		{"invalid channel id", types.Counterparty{"portidone", "(InvalidChannel)"}, false},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.counterparty.ValidateBasic()
		if tc.expPass {
			require.NoError(t, err, "valid test case %d failed: %s", i, tc.name)
		} else {
			require.Error(t, err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

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
