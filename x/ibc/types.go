package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	wire "github.com/cosmos/cosmos-sdk/wire"

	types "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// ----------------------------------
// IBCReceiveMsg

// IBCReceiveMsg defines the message that a relayer uses to post an IBCPacket
// to the destination chain.
type ReceiveMsg struct {
	types.Packet
	Relayer  sdk.Address
	Sequence int64
}

func (msg ReceiveMsg) Type() string {
	return "ibc"
}

func (msg ReceiveMsg) Get(key interface{}) interface{} {
	return nil
}

func (msg ReceiveMsg) GetSignBytes() []byte {
	cdc := wire.NewCodec()
	bz, err := cdc.MarshalBinary(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg ReceiveMsg) ValidateBasic() sdk.Error {
	return msg.Packet.ValidateBasic()
}

// x/bank/tx.go SendMsg.GetSigners()
func (msg ReceiveMsg) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Relayer}
}
