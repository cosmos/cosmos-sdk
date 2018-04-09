package ibc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestIBCReceiveMsgType(t *testing.T) {
	packet := constructIBCPacket(true)
	msg := constructReceiveMsg(packet)
	assert.Equal(t, msg.Type(), "ibc")
}

func TestIBCReceiveMsgValidation(t *testing.T) {
	validPacket := constructIBCPacket(true)
	invalidPacket := constructIBCPacket(false)

	cases := []struct {
		valid bool
		msg   ReceiveMsg
	}{
		{true, constructReceiveMsg(validPacket)},
		{false, constructReceiveMsg(invalidPacket)},
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

func TestIBCReceiveMsgSigners(t *testing.T) {
	packet := constructIBCPacket(true)
	relayer := sdk.Address([]byte("relayer"))
	msg := ReceiveMsg{
		Packet:   packet,
		Relayer:  relayer,
		Sequence: 0,
	}
	assert.Equal(t, []sdk.Address{relayer}, msg.GetSigners())
}

// -------------------------------
// Helpers

type myPayload struct {
	valid bool
}

func (p myPayload) Type() string {
	return "my"
}

func (p myPayload) ValidateBasic() sdk.Error {
	if p.valid {
		return nil
	}
	return sdk.NewError(42, "")
}

func (p myPayload) GetSigners() []sdk.Address {
	return []sdk.Address{}
}

func constructIBCPacket(valid bool) Packet {
	srcChain := "source-chain"
	destChain := "dest-chain"

	return Packet{
		Payload:   myPayload{valid},
		SrcChain:  srcChain,
		DestChain: destChain,
	}
}

func constructReceiveMsg(packet Packet) ReceiveMsg {
	return ReceiveMsg{
		Packet:   packet,
		Relayer:  sdk.Address([]byte("relayer")),
		Sequence: 0,
	}

}
