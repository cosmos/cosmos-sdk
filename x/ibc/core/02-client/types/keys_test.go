package types_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
)

// tests ParseClientIdentifier and IsValidClientID
func TestParseClientIdentifier(t *testing.T) {
	testCases := []struct {
		name       string
		clientID   string
		clientType string
		expSeq     uint64
		expPass    bool
	}{
		{"valid 0", "tendermint-0", "tendermint", 0, true},
		{"valid 1", "tendermint-1", "tendermint", 1, true},
		{"valid solemachine", "solomachine-v1-1", "solomachine-v1", 1, true},
		{"valid large sequence", types.FormatClientIdentifier("tendermint", math.MaxUint64), "tendermint", math.MaxUint64, true},
		{"valid short client type", "t-0", "t", 0, true},
		// one above uint64 max
		{"invalid uint64", "tendermint-18446744073709551616", "tendermint", 0, false},
		// uint64 == 20 characters
		{"invalid large sequence", "tendermint-2345682193567182931243", "tendermint", 0, false},
		{"missing dash", "tendermint0", "tendermint", 0, false},
		{"blank id", "               ", "    ", 0, false},
		{"empty id", "", "", 0, false},
		{"negative sequence", "tendermint--1", "tendermint", 0, false},
		{"invalid format", "tendermint-tm", "tendermint", 0, false},
		{"empty clientype", " -100", "tendermint", 0, false},
	}

	for _, tc := range testCases {

		clientType, seq, err := types.ParseClientIdentifier(tc.clientID)
		valid := types.IsValidClientID(tc.clientID)
		require.Equal(t, tc.expSeq, seq, tc.clientID)

		if tc.expPass {
			require.NoError(t, err, tc.name)
			require.True(t, valid)
			require.Equal(t, tc.clientType, clientType)
		} else {
			require.Error(t, err, tc.name, tc.clientID)
			require.False(t, valid)
			require.Equal(t, "", clientType)
		}
	}
}
