package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestFungibleTokenPacketDataValidateBasic tests ValidateBasic for FungibleTokenPacketData
func TestFungibleTokenPacketDataValidateBasic(t *testing.T) {
	testPacketDataTransfer := []FungibleTokenPacketData{
		NewFungibleTokenPacketData(coins, addr1, addr2, true),             // valid msg
		NewFungibleTokenPacketData(invalidDenomCoins, addr1, addr2, true), // invalid amount
		NewFungibleTokenPacketData(negativeCoins, addr1, addr2, false),    // amount contains negative coin
		NewFungibleTokenPacketData(coins, emptyAddr, addr2, false),        // missing sender address
		NewFungibleTokenPacketData(coins, addr1, emptyAddr, false),        // missing recipient address
	}

	testCases := []struct {
		packetData FungibleTokenPacketData
		expPass    bool
		errMsg     string
	}{
		{testPacketDataTransfer[0], true, ""},
		{testPacketDataTransfer[1], false, "invalid amount"},
		{testPacketDataTransfer[2], false, "amount contains negative coin"},
		{testPacketDataTransfer[3], false, "missing sender address"},
		{testPacketDataTransfer[4], false, "missing recipient address"},
	}

	for i, tc := range testCases {
		err := tc.packetData.ValidateBasic()
		if tc.expPass {
			require.NoError(t, err, "PacketDataTransfer %d failed: %v", i, err)
		} else {
			require.Error(t, err, "Invalid PacketDataTransfer %d passed: %s", i, tc.errMsg)
		}
	}
}
