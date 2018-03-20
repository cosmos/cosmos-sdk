package ibc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibc "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

func TestIBCReceiveMsgValidation(t *testing.T) {
	validPacket := constructIBCPacket(true)
	invalidPacket := constructIBCPacket(false)

	cases := []struct {
		valid bool
		msg   ReceiveMsg
	}{
		{true, ReceiveMsg{validPacket, sdk.Address([]byte("relayer")), 0}},
		{false, ReceiveMsg{invalidPacket, sdk.Address([]byte("relayer")), 0}},
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

func constructIBCPacket(valid bool) ibc.Packet {
	srcChain := "source-chain"
	destChain := "dest-chain"

	return ibc.Packet{
		Payload:   myPayload{valid},
		SrcChain:  srcChain,
		DestChain: destChain,
	}
}
