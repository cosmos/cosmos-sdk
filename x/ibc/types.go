package ibc

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"
)

var (
	msgCdc *wire.Codec
)

func init() {
	msgCdc = wire.NewCodec()
}

// ------------------------------
// IBCPacket

// nolint - TODO rename to Packet as IBCPacket stutters (golint)
// IBCPacket defines a piece of data that can be send between two separate
// blockchains.
type IBCPacket struct {
	SrcAddr   sdk.AccAddress
	DestAddr  sdk.AccAddress
	Coins     sdk.Coins
	SrcChain  string
	DestChain string
}

func NewIBCPacket(srcAddr sdk.AccAddress, destAddr sdk.AccAddress, coins sdk.Coins,
	srcChain string, destChain string) IBCPacket {

	return IBCPacket{
		SrcAddr:   srcAddr,
		DestAddr:  destAddr,
		Coins:     coins,
		SrcChain:  srcChain,
		DestChain: destChain,
	}
}

//nolint
func (p IBCPacket) GetSignBytes() []byte {
	b, err := msgCdc.MarshalJSON(p)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

// validator the ibc packey
func (p IBCPacket) ValidateBasic() sdk.Error {
	if p.SrcChain == p.DestChain {
		return ErrIdenticalChains(DefaultCodespace).TraceSDK("")
	}
	if !p.Coins.IsValid() {
		return sdk.ErrInvalidCoins("")
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
func (msg IBCTransferMsg) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{msg.SrcAddr} }

// get the sign bytes for ibc transfer message
func (msg IBCTransferMsg) GetSignBytes() []byte {
	return msg.IBCPacket.GetSignBytes()
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
	Relayer  sdk.AccAddress
	Sequence int64
}

// nolint
func (msg IBCReceiveMsg) Type() string             { return "ibc" }
func (msg IBCReceiveMsg) ValidateBasic() sdk.Error { return msg.IBCPacket.ValidateBasic() }

// x/bank/tx.go MsgSend.GetSigners()
func (msg IBCReceiveMsg) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{msg.Relayer} }

// get the sign bytes for ibc receive message
func (msg IBCReceiveMsg) GetSignBytes() []byte {
	b, err := msgCdc.MarshalJSON(struct {
		IBCPacket json.RawMessage
		Relayer   sdk.AccAddress
		Sequence  int64
	}{
		IBCPacket: json.RawMessage(msg.IBCPacket.GetSignBytes()),
		Relayer:   msg.Relayer,
		Sequence:  msg.Sequence,
	})
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}
