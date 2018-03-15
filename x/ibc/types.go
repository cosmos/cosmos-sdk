package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	wire "github.com/cosmos/cosmos-sdk/wire"
)

type IBCPacket struct {
	SrcAddr   sdk.Address
	DestAddr  sdk.Address
	Coins     sdk.Coins
	SrcChain  string
	DestChain string
}

func newCodec() *wire.Codec {
	return wire.NewCodec()
}

type IBCTransferMsg struct {
	IBCPacket
}

func (msg IBCTransferMsg) Type() string {
	return "ibctransfer"
}

func (msg IBCTransferMsg) Get(key interface{}) interface{} {
	return nil
}

func (msg IBCTransferMsg) GetSignBytes() []byte {
	cdc := newCodec()
	bz, err := cdc.MarshalBinary(msg.IBCPacket)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg IBCTransferMsg) ValidateBasic() sdk.Error {
	return nil
}

// x/bank/tx.go SendMsg.GetSigners()
func (msg IBCTransferMsg) GetSigners() []sdk.Address {
	return []sdk.Address{msg.SrcAddr}
}

type IBCReceiveMsg struct {
	IBCPacket
	Relayer  sdk.Address
	Sequence int64
}

func (msg IBCReceiveMsg) Type() string {
	return "ibcreceive"
}

func (msg IBCReceiveMsg) Get(key interface{}) interface{} {
	return nil
}

func (msg IBCReceiveMsg) GetSignBytes() []byte {
	cdc := newCodec()
	bz, err := cdc.MarshalBinary(msg.IBCPacket)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg IBCReceiveMsg) ValidateBasic() sdk.Error {
	return nil
}

// x/bank/tx.go SendMsg.GetSigners()
func (msg IBCReceiveMsg) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Relayer}
}
