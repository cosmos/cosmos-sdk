package ibc

import (
	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"
)

// ------------------------------
// IBCPacket

// nolint - TODO rename to Packet as IBCPacket stutters (golint)
// IBCPacket defines a piece of data that can be send between two separate
// blockchains.
type IBCPacket struct {
	SrcAddr   bam.Address
	DestAddr  bam.Address
	Coins     bam.Coins
	SrcChain  string
	DestChain string
}

func NewIBCPacket(srcAddr bam.Address, destAddr bam.Address, coins bam.Coins,
	srcChain string, destChain string) IBCPacket {

	return IBCPacket{
		SrcAddr:   srcAddr,
		DestAddr:  destAddr,
		Coins:     coins,
		SrcChain:  srcChain,
		DestChain: destChain,
	}
}

// validator the ibc packey
func (ibcp IBCPacket) ValidateBasic() sdk.Error {
	if ibcp.SrcChain == ibcp.DestChain {
		return ErrIdenticalChains(DefaultCodespace).Trace("")
	}
	if !ibcp.Coins.IsValid() {
		return bam.ErrInvalidCoins("")
	}
	return nil
}

// ----------------------------------
// IBCTransferMsg

// nolint - TODO rename to TransferMsg as folks will reference with ibc.TransferMsg
// IBCTransferMsg defines how another module can send an IBCPacket.
type IBCTransferMsg struct {
	IBCPacket
}

// nolint
func (msg IBCTransferMsg) Type() string { return "ibc" }

// x/bank/tx.go MsgSend.GetSigners()
func (msg IBCTransferMsg) GetSigners() []bam.Address { return []bam.Address{msg.SrcAddr} }

// get the sign bytes for ibc transfer message
func (msg IBCTransferMsg) GetSignBytes() []byte {
	cdc := wire.NewCodec()
	bz, err := cdc.MarshalBinary(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

// validate ibc transfer message
func (msg IBCTransferMsg) ValidateBasic() sdk.Error {
	return msg.IBCPacket.ValidateBasic()
}

// ----------------------------------
// IBCReceiveMsg

// nolint - TODO rename to ReceiveMsg as folks will reference with ibc.ReceiveMsg
// IBCReceiveMsg defines the message that a relayer uses to post an IBCPacket
// to the destination chain.
type IBCReceiveMsg struct {
	IBCPacket
	Relayer  bam.Address
	Sequence int64
}

// nolint
func (msg IBCReceiveMsg) Type() string             { return "ibc" }
func (msg IBCReceiveMsg) ValidateBasic() sdk.Error { return msg.IBCPacket.ValidateBasic() }

// x/bank/tx.go MsgSend.GetSigners()
func (msg IBCReceiveMsg) GetSigners() []bam.Address { return []bam.Address{msg.Relayer} }

// get the sign bytes for ibc receive message
func (msg IBCReceiveMsg) GetSignBytes() []byte {
	cdc := wire.NewCodec()
	bz, err := cdc.MarshalBinary(msg)
	if err != nil {
		panic(err)
	}
	return bz
}
