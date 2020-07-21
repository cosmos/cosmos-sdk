package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

func TestPacketValidateBasic(t *testing.T) {
	testCases := []struct {
		packet  types.Packet
		expPass bool
		errMsg  string
	}{
		{types.NewPacket(validPacketData, 1, portid, chanid, cpportid, cpchanid, timeoutEpoch, timeoutHeight, timeoutTimestamp), true, ""},
		{types.NewPacket(validPacketData, 0, portid, chanid, cpportid, cpchanid, timeoutEpoch, timeoutHeight, timeoutTimestamp), false, "invalid sequence"},
		{types.NewPacket(validPacketData, 1, invalidPort, chanid, cpportid, cpchanid, timeoutEpoch, timeoutHeight, timeoutTimestamp), false, "invalid source port"},
		{types.NewPacket(validPacketData, 1, portid, invalidChannel, cpportid, cpchanid, timeoutEpoch, timeoutHeight, timeoutTimestamp), false, "invalid source channel"},
		{types.NewPacket(validPacketData, 1, portid, chanid, invalidPort, cpchanid, timeoutEpoch, timeoutHeight, timeoutTimestamp), false, "invalid destination port"},
		{types.NewPacket(validPacketData, 1, portid, chanid, cpportid, invalidChannel, timeoutEpoch, timeoutHeight, timeoutTimestamp), false, "invalid destination channel"},
		{types.NewPacket(validPacketData, 1, portid, chanid, cpportid, cpchanid, timeoutEpoch, disabledTimeout, disabledTimeout), false, "disabled both timeout height and timestamp"},
		{types.NewPacket(validPacketData, 1, portid, chanid, cpportid, cpchanid, timeoutEpoch, disabledTimeout, timeoutTimestamp), true, "disabled timeout height, valid timeout timestamp"},
		{types.NewPacket(validPacketData, 1, portid, chanid, cpportid, cpchanid, timeoutEpoch, timeoutHeight, disabledTimeout), true, "disabled timeout timestamp, valid timeout height"},
		{types.NewPacket(unknownPacketData, 1, portid, chanid, cpportid, cpchanid, timeoutEpoch, timeoutHeight, timeoutTimestamp), true, ""},
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
