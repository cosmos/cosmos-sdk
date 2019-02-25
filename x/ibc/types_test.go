package ibc

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// --------------------------------
// IBCPacket Tests

func TestIBCPacketValidation(t *testing.T) {
	cases := []struct {
		valid  bool
		packet IBCPacket
	}{
		{true, constructIBCPacket(true)},
		{false, constructIBCPacket(false)},
	}

	for i, tc := range cases {
		err := tc.packet.ValidateBasic()
		if tc.valid {
			require.Nil(t, err, "%d: %+v", i, err)
		} else {
			require.NotNil(t, err, "%d", i)
		}
	}
}

// -------------------------------
// MsgIBCTransfer Tests

func TestIBCTransferMsg(t *testing.T) {
	packet := constructIBCPacket(true)
	msg := MsgIBCTransfer{packet}

	require.Equal(t, msg.Route(), "ibc")
}

func TestIBCTransferMsgValidation(t *testing.T) {
	validPacket := constructIBCPacket(true)
	invalidPacket := constructIBCPacket(false)

	cases := []struct {
		valid bool
		msg   MsgIBCTransfer
	}{
		{true, MsgIBCTransfer{validPacket}},
		{false, MsgIBCTransfer{invalidPacket}},
	}

	for i, tc := range cases {
		err := tc.msg.ValidateBasic()
		if tc.valid {
			require.Nil(t, err, "%d: %+v", i, err)
		} else {
			require.NotNil(t, err, "%d", i)
		}
	}
}

// -------------------------------
// MsgIBCReceive Tests

func TestIBCReceiveMsg(t *testing.T) {
	packet := constructIBCPacket(true)
	msg := MsgIBCReceive{packet, sdk.AccAddress([]byte("relayer")), 0}

	require.Equal(t, msg.Route(), "ibc")
}

func TestIBCReceiveMsgValidation(t *testing.T) {
	validPacket := constructIBCPacket(true)
	invalidPacket := constructIBCPacket(false)

	cases := []struct {
		valid bool
		msg   MsgIBCReceive
	}{
		{true, MsgIBCReceive{validPacket, sdk.AccAddress([]byte("relayer")), 0}},
		{false, MsgIBCReceive{invalidPacket, sdk.AccAddress([]byte("relayer")), 0}},
	}

	for i, tc := range cases {
		err := tc.msg.ValidateBasic()
		if tc.valid {
			require.Nil(t, err, "%d: %+v", i, err)
		} else {
			require.NotNil(t, err, "%d", i)
		}
	}
}

// -------------------------------
// Helpers

func constructIBCPacket(valid bool) IBCPacket {
	srcAddr := sdk.AccAddress([]byte("source"))
	destAddr := sdk.AccAddress([]byte("destination"))
	coins := sdk.Coins{sdk.NewInt64Coin("atom", 10)}
	srcChain := "source-chain"
	destChain := "dest-chain"

	if valid {
		return NewIBCPacket(srcAddr, destAddr, coins, srcChain, destChain)
	}
	return NewIBCPacket(srcAddr, destAddr, coins, srcChain, srcChain)
}
