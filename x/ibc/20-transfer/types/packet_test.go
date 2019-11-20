package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestPacketDataValidation tests ValidateBasic for PacketData
func TestPacketDataValidation(t *testing.T) {
	testPacketData := []PacketData{
		NewPacketData(coins, addr1, addr2, true),             // valid msg
		NewPacketData(invalidDenomCoins, addr1, addr2, true), // invalid amount
		NewPacketData(negativeCoins, addr1, addr2, false),    // amount contains negative coin
		NewPacketData(coins, emptyAddr, addr2, false),        // missing sender address
		NewPacketData(coins, addr1, emptyAddr, false),        // missing recipient address
	}

	testCases := []struct {
		packetData PacketData
		expPass    bool
		errMsg     string
	}{
		{testPacketData[0], true, ""},
		{testPacketData[1], false, "invalid amount"},
		{testPacketData[2], false, "amount contains negative coin"},
		{testPacketData[3], false, "missing sender address"},
		{testPacketData[4], false, "missing recipient address"},
	}

	for i, tc := range testCases {
		err := tc.packetData.ValidateBasic()
		if tc.expPass {
			require.NoError(t, err, "PacketData %d failed: %v", i, err)
		} else {
			require.Error(t, err, "Invalid PacketData %d passed: %s", i, tc.errMsg)
		}
	}
}
