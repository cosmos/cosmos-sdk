package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestFungibleTokenPacketDataValidateBasic tests ValidateBasic for FungibleTokenPacketData
func TestFungibleTokenPacketDataValidateBasic(t *testing.T) {
	testCases := []struct {
		name       string
		packetData FungibleTokenPacketData
		expPass    bool
	}{
		{"valid packet", NewFungibleTokenPacketData(coin, addr1.String(), addr2, true), true},
		{"invalid amount", NewFungibleTokenPacketData(invalidDenomCoin, addr1.String(), addr2, true), false},
		{"amount contains negative coin", NewFungibleTokenPacketData(negativeCoin, addr1.String(), addr2, true), false},
		{"missing sender address", NewFungibleTokenPacketData(coin, emptyAddr.String(), addr2, false), false},
		{"missing recipient address", NewFungibleTokenPacketData(coin, addr1.String(), emptyAddr.String(), false), false},
	}

	for i, tc := range testCases {
		err := tc.packetData.ValidateBasic()
		if tc.expPass {
			require.NoError(t, err, "valid test case %d failed: %v", i, err)
		} else {
			require.Error(t, err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}
