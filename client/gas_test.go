package client_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
)

func TestParseGasSetting(t *testing.T) {
	testCases := []struct {
		name      string
		input     string
		expected  client.GasSetting
		expectErr bool
	}{
		{"empty input", "", client.DefaultGasSetting, false},
		{"auto", flags.GasFlagAuto, client.GasSetting{true, 0}, false},
		{"valid custom gas", "73800", client.GasSetting{false, 73800}, false},
		{"invalid custom gas", "-73800", client.GasSetting{false, 0}, true},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			gs, err := client.ParseGasSetting(tc.input)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, gs)
			}
		})
	}
}
