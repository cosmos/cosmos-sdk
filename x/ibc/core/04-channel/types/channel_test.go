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
