package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	wire "github.com/cosmos/cosmos-sdk/wire"
)

// ------------------------------
// Msg
// Msg defines inter-blockchain Msg
// that can be proved by light-client protocol

type Msg interface {
	Type() string
	GetSigners() []sdk.Address
	ValidateBasic() sdk.Error
}

// ------------------------------
// Packet

// Packet defines a piece of data that can be send between two separate
// blockchains.
type Packet struct {
	Msg       Msg
	SrcChain  string
	DestChain string
}

/*
func NewIBCPacket(srcAddr sdk.Address, destAddr sdk.Address, coins sdk.Coins,
	srcChain string, destChain string) IBCPacket {

	return Packet{
		SrcAddr:   srcAddr,
		DestAddr:  destAddr,
		Coins:     coins,
		SrcChain:  srcChain,
		DestChain: destChain,
	}
}
*/
func (packet Packet) ValidateBasic() sdk.Error {
	if packet.SrcChain == packet.DestChain {
		return ErrIdenticalChains().Trace("")
	}
	return packet.Msg.ValidateBasic()
}

// ----------------------------------
// IBCSendMsg

// IBCSendMsg defines how another module can send an IBCPacket.
type IBCSendMsg struct {
	Packet
}

func (msg IBCSendMsg) Type() string {
	return "ibc"
}

func (msg IBCSendMsg) Get(key interface{}) interface{} {
	return nil
}

func (msg IBCSendMsg) GetSignBytes() []byte {
	cdc := wire.NewCodec()
	bz, err := cdc.MarshalBinary(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg IBCSendMsg) ValidateBasic() sdk.Error {
	return msg.Packet.ValidateBasic()
}

// x/bank/tx.go SendMsg.GetSigners()
func (msg IBCSendMsg) GetSigners() []sdk.Address {
	return msg.Packet.Msg.GetSigners()
}

// ----------------------------------
// IBCReceiveMsg

// IBCReceiveMsg defines the message that a relayer uses to post an IBCPacket
// to the destination chain.
type IBCReceiveMsg struct {
	Packet
	Relayer  sdk.Address
	Sequence int64
}

func (msg IBCReceiveMsg) Type() string {
	return "ibc"
}

func (msg IBCReceiveMsg) Get(key interface{}) interface{} {
	return nil
}

func (msg IBCReceiveMsg) GetSignBytes() []byte {
	cdc := wire.NewCodec()
	bz, err := cdc.MarshalBinary(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg IBCReceiveMsg) ValidateBasic() sdk.Error {
	return msg.Packet.ValidateBasic()
}

// x/bank/tx.go SendMsg.GetSigners()
func (msg IBCReceiveMsg) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Relayer}
}
