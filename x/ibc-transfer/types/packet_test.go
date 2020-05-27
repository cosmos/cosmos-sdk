package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestFungibleTokenPacketDataValidateBasic tests ValidateBasic for FungibleTokenPacketData
func TestFungibleTokenPacketDataValidateBasic(t *testing.T) {
	testPacketDataTransfer := []FungibleTokenPacketData{
		NewFungibleTokenPacketData(coins, addr1.String(), addr2),              // valid msg
		NewFungibleTokenPacketData(invalidDenomCoins, addr1.String(), addr2),  // invalid amount
		NewFungibleTokenPacketData(negativeCoins, addr1.String(), addr2),      // amount contains negative coin
		NewFungibleTokenPacketData(coins, emptyAddr.String(), addr2),          // missing sender address
		NewFungibleTokenPacketData(coins, addr1.String(), emptyAddr.String()), // missing recipient address
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
