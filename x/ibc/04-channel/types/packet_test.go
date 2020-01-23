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
		{NewPacket(validPacketT{}, 1, portid, chanid, cpportid, cpchanid), true, ""},
		{NewPacket(validPacketT{}, 0, portid, chanid, cpportid, cpchanid), false, "invalid sequence"},
		{NewPacket(validPacketT{}, 1, invalidPort, chanid, cpportid, cpchanid), false, "invalid source port"},
		{NewPacket(validPacketT{}, 1, portid, invalidChannel, cpportid, cpchanid), false, "invalid source channel"},
		{NewPacket(validPacketT{}, 1, portid, chanid, invalidPort, cpchanid), false, "invalid destination port"},
		{NewPacket(validPacketT{}, 1, portid, chanid, cpportid, invalidChannel), false, "invalid destination channel"},
		{NewPacket(invalidPacketT{}, 1, portid, chanid, cpportid, cpchanid), false, "invalid packet data"},
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
