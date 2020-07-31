package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	denom  = "transfer/gaiachannel/atom"
	amount = uint64(100)
)

// TestFungibleTokenPacketDataValidateBasic tests ValidateBasic for FungibleTokenPacketData
func TestFungibleTokenPacketDataValidateBasic(t *testing.T) {
	testCases := []struct {
		name       string
		packetData FungibleTokenPacketData
		expPass    bool
	}{
		{"valid packet", NewFungibleTokenPacketData(denom, amount, addr1.String(), addr2), true},
		{"invalid denom", NewFungibleTokenPacketData("", amount, addr1.String(), addr2), false},
		{"invalid amount", NewFungibleTokenPacketData(denom, 0, addr1.String(), addr2), false},
		{"missing sender address", NewFungibleTokenPacketData(denom, amount, emptyAddr.String(), addr2), false},
		{"missing recipient address", NewFungibleTokenPacketData(denom, amount, addr1.String(), emptyAddr.String()), false},
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
