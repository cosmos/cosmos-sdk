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
	bz, err := cdc.MarshalBinary(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg IBCTransferMsg) ValidateBasic() sdk.Error {
	return msg.Coins.ValidateBasic()
}

// x/bank/tx.go SendMsg.GetSigners()
func (msg IBCTransferMsg) GetSigners() []sdk.Address {
	addrs := []sdk.Address{msg.SrcAddr}
}

type IBCReceiveMsg struct {
	IBCPacket
	Relayer sdk.Address
}
