package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPacketValidateBasic(t *testing.T) {
	testCases := []struct {
		packet  Packet
		expPass bool
		errMsg  string
	}{
		{NewPacket(validPacketData, 1, portid, chanid, cpportid, cpchanid, timeoutHeight, timeoutTimestamp), true, ""},
		{NewPacket(validPacketData, 0, portid, chanid, cpportid, cpchanid, timeoutHeight, timeoutTimestamp), false, "invalid sequence"},
		{NewPacket(validPacketData, 1, invalidPort, chanid, cpportid, cpchanid, timeoutHeight, timeoutTimestamp), false, "invalid source port"},
		{NewPacket(validPacketData, 1, portid, invalidChannel, cpportid, cpchanid, timeoutHeight, timeoutTimestamp), false, "invalid source channel"},
		{NewPacket(validPacketData, 1, portid, chanid, invalidPort, cpchanid, timeoutHeight, timeoutTimestamp), false, "invalid destination port"},
		{NewPacket(validPacketData, 1, portid, chanid, cpportid, invalidChannel, timeoutHeight, timeoutTimestamp), false, "invalid destination channel"},
		{NewPacket(validPacketData, 1, portid, chanid, cpportid, cpchanid, disabledTimeout, disabledTimeout), false, "disabled both timeout height and timestamp"},
		{NewPacket(validPacketData, 1, portid, chanid, cpportid, cpchanid, disabledTimeout, timeoutTimestamp), true, "disabled timeout height, valid timeout timestamp"},
		{NewPacket(validPacketData, 1, portid, chanid, cpportid, cpchanid, timeoutHeight, disabledTimeout), true, "disabled timeout timestamp, valid timeout height"},
		{NewPacket(unknownPacketData, 1, portid, chanid, cpportid, cpchanid, timeoutHeight, timeoutTimestamp), true, ""},
	}

	for i, tc := range testCases {
		err := tc.packet.ValidateBasic()
		if tc.expPass {
			require.NoError(t, err, "Msg %d failed: %s", i, tc.errMsg)
		} else {
			require.Error(t, err, "Invalid Msg %d passed: %s", i, tc.errMsg)
		}
	}
}
