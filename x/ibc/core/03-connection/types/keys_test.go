package types_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/types"
	"github.com/stretchr/testify/require"
)

// tests ParseConnectionSequence and IsValidConnectionID
func TestParseConnectionSequence(t *testing.T) {
	testCases := []struct {
		name         string
		connectionID string
		expSeq       uint64
		expPass      bool
	}{
		{"valid 0", "connection-0", 0, true},
		{"valid 1", "connection-1", 1, true},
		{"valid large sequence", "connection-234568219356718293", 234568219356718293, true},
		// uint64 == 20 characters
		{"invalid large sequence", "connection-2345682193567182931243", 0, false},
		{"capital prefix", "Connection-0", 0, false},
		{"missing dash", "connection0", 0, false},
		{"blank id", "               ", 0, false},
		{"empty id", "", 0, false},
		{"negative sequence", "connection--1", 0, false},
	}

	for _, tc := range testCases {

		seq, err := types.ParseConnectionSequence(tc.connectionID)
		valid := types.IsValidConnectionID(tc.connectionID)
		require.Equal(t, tc.expSeq, seq)

		if tc.expPass {
			require.NoError(t, err, tc.name)
			require.True(t, valid)
		} else {
			require.Error(t, err, tc.name)
			require.False(t, valid)
		}
	}
}
