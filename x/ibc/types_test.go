package ibc

import (
/*	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibc "github.com/cosmos/cosmos-sdk/x/ibc/types"
*/
)

/*
// --------------------------------
// IBCPacket Tests

func TestIBCPacketValidation(t *testing.T) {
	cases := []struct {
		valid  bool
		packet ibc.Packet
	}{
		{true, constructIBCPacket(true)},
		{false, constructIBCPacket(false)},
	}

	for i, tc := range cases {
		err := tc.packet.ValidateBasic()
		if tc.valid {
			assert.Nil(t, err, "%d: %+v", i, err)
		} else {
			assert.NotNil(t, err, "%d", i)
		}
	}
}
*/
/*
// -------------------------------
// IBCTransferMsg Tests

func TestIBCTransferMsg(t *testing.T) {
	packet := constructIBCPacket(true)
	msg := IBCTransferMsg{packet}

	assert.Equal(t, msg.Type(), "ibc")
}

func TestIBCTransferMsgValidation(t *testing.T) {
	validPacket := constructIBCPacket(true)
	invalidPacket := constructIBCPacket(false)

	cases := []struct {
		valid bool
		msg   IBCTransferMsg
	}{
		{true, IBCTransferMsg{validPacket}},
		{false, IBCTransferMsg{invalidPacket}},
	}

	for i, tc := range cases {
		err := tc.msg.ValidateBasic()
		if tc.valid {
			assert.Nil(t, err, "%d: %+v", i, err)
		} else {
			assert.NotNil(t, err, "%d", i)
		}
	}
}
*/
/*
// -------------------------------
// IBCReceiveMsg Tests

func TestIBCReceiveMsg(t *testing.T) {
	packet := constructIBCPacket(true)
	msg := ReceiveMsg{packet, sdk.Address([]byte("relayer")), 0}

	assert.Equal(t, msg.Type(), "ibc")
}
*/
/*
func TestIBCReceiveMsgValidation(t *testing.T) {
	validPacket := constructIBCPacket(true)
	invalidPacket := constructIBCPacket(false)

	cases := []struct {
		valid bool
		msg   IBCReceiveMsg
	}{
		{true, IBCReceiveMsg{validPacket, sdk.Address([]byte("relayer")), 0}},
		{false, IBCReceiveMsg{invalidPacket, sdk.Address([]byte("relayer")), 0}},
	}

	for i, tc := range cases {
		err := tc.msg.ValidateBasic()
		if tc.valid {
			assert.Nil(t, err, "%d: %+v", i, err)
		} else {
			assert.NotNil(t, err, "%d", i)
		}
	}
}
*/
// -------------------------------
// Helpers
/*
func constructIBCPacket(valid bool) ibc.Packet {
	if !valid {
		panic("No invalid packet generation for now")
	}

	srcAddr := sdk.Address([]byte("source"))
	destAddr := sdk.Address([]byte("destination"))
	coins := sdk.Coins{{"atom", 10}}
	srcChain := "source-chain"
	destChain := "dest-chain"

	if valid {
		return NewIBCPacket(srcAddr, destAddr, coins, srcChain, destChain)
	} else {
		return NewIBCPacket(srcAddr, destAddr, coins, srcChain, srcChain)
	}
}*/
