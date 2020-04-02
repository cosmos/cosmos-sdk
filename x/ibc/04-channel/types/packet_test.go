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
		{NewPacket(validPacketData, 1, portid, chanid, cpportid, cpchanid, timeout), true, ""},
		{NewPacket(validPacketData, 0, portid, chanid, cpportid, cpchanid, timeout), false, "invalid sequence"},
		{NewPacket(validPacketData, 1, invalidPort, chanid, cpportid, cpchanid, timeout), false, "invalid source port"},
		{NewPacket(validPacketData, 1, portid, invalidChannel, cpportid, cpchanid, timeout), false, "invalid source channel"},
		{NewPacket(validPacketData, 1, portid, chanid, invalidPort, cpchanid, timeout), false, "invalid destination port"},
		{NewPacket(validPacketData, 1, portid, chanid, cpportid, invalidChannel, timeout), false, "invalid destination channel"},
		{NewPacket(unknownPacketData, 1, portid, chanid, cpportid, cpchanid, timeout), true, ""},
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
