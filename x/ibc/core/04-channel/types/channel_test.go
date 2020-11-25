package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
)

func TestChannelValidateBasic(t *testing.T) {
	counterparty := types.Counterparty{"portidone", "channelidone"}
	testCases := []struct {
		name    string
		channel types.Channel
		expPass bool
	}{
		{"valid channel", types.NewChannel(types.TRYOPEN, types.ORDERED, counterparty, connHops, version), true},
		{"invalid state", types.NewChannel(types.UNINITIALIZED, types.ORDERED, counterparty, connHops, version), false},
		{"invalid order", types.NewChannel(types.TRYOPEN, types.NONE, counterparty, connHops, version), false},
		{"more than 1 connection hop", types.NewChannel(types.TRYOPEN, types.ORDERED, counterparty, []string{"connection1", "connection2"}, version), false},
		{"invalid connection hop identifier", types.NewChannel(types.TRYOPEN, types.ORDERED, counterparty, []string{"(invalid)"}, version), false},
		{"invalid counterparty", types.NewChannel(types.TRYOPEN, types.ORDERED, types.NewCounterparty("(invalidport)", "channelidone"), connHops, version), false},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.channel.ValidateBasic()
		if tc.expPass {
			require.NoError(t, err, "valid test case %d failed: %s", i, tc.name)
		} else {
			require.Error(t, err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

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
