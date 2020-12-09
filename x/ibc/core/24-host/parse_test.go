package host_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
)

func TestParseIdentifier(t *testing.T) {
	testCases := []struct {
		name       string
		identifier string
		prefix     string
		expSeq     uint64
		expPass    bool
	}{
		{"valid 0", "connection-0", "connection-", 0, true},
		{"valid 1", "connection-1", "connection-", 1, true},
		{"valid large sequence", connectiontypes.FormatConnectionIdentifier(math.MaxUint64), "connection-", math.MaxUint64, true},
		// one above uint64 max
		{"invalid uint64", "connection-18446744073709551616", "connection-", 0, false},
		// uint64 == 20 characters
		{"invalid large sequence", "connection-2345682193567182931243", "connection-", 0, false},
		{"capital prefix", "Connection-0", "connection-", 0, false},
		{"double prefix", "connection-connection-0", "connection-", 0, false},
		{"doesn't have prefix", "connection-0", "prefix", 0, false},
		{"missing dash", "connection0", "connection-", 0, false},
		{"blank id", "               ", "connection-", 0, false},
		{"empty id", "", "connection-", 0, false},
		{"negative sequence", "connection--1", "connection-", 0, false},
	}

	for _, tc := range testCases {

		seq, err := host.ParseIdentifier(tc.identifier, tc.prefix)
		require.Equal(t, tc.expSeq, seq)

		if tc.expPass {
			require.NoError(t, err, tc.name)
		} else {
			require.Error(t, err, tc.name)
		}
	}
}
