package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestPacketDataTransferValidation tests ValidateBasic for PacketDataTransfer
func TestPacketDataTransferValidation(t *testing.T) {
	testPacketDataTransfer := []PacketDataTransfer{
		NewPacketDataTransfer(coins, addr1, addr2, true, 100),             // valid msg
		NewPacketDataTransfer(invalidDenomCoins, addr1, addr2, true, 100), // invalid amount
		NewPacketDataTransfer(negativeCoins, addr1, addr2, false, 100),    // amount contains negative coin
		NewPacketDataTransfer(coins, emptyAddr, addr2, false, 100),        // missing sender address
		NewPacketDataTransfer(coins, addr1, emptyAddr, false, 100),        // missing recipient address
	}

	testCases := []struct {
		packetData PacketDataTransfer
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
