package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCounterpartyValidateBasic(t *testing.T) {
	testCases := []struct {
		name         string
		counterparty Counterparty
		expPass      bool
	}{
		{"valid counterparty", Counterparty{"portidone", "channelidone"}, true},
		{"invalid port id", Counterparty{"(InvalidPort)", "channelidone"}, false},
		{"invalid channel id", Counterparty{"portidone", "(InvalidChannel)"}, false},
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
