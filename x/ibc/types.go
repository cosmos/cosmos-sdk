package ibc

import (
	"encoding/json"

	codec "github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	msgCdc *codec.Codec
)

func init() {
	msgCdc = codec.New()
}

// ------------------------------
// IBCPacket

// nolint - TODO rename to Packet as IBCPacket stutters (golint)
// IBCPacket defines a piece of data that can be send between two separate
// blockchains.
type IBCPacket struct {
	SrcAddr   sdk.AccAddress `json:"src_addr"`
	DestAddr  sdk.AccAddress `json:"dest_addr"`
	Coins     sdk.Coins      `json:"coins"`
	SrcChain  string         `json:"src_chain"`
	DestChain string         `json:"dest_chain"`
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
// MsgIBCTransfer

// nolint - TODO rename to TransferMsg as folks will reference with ibc.TransferMsg
// MsgIBCTransfer defines how another module can send an IBCPacket.
type MsgIBCTransfer struct {
	IBCPacket
}

// nolint
func (msg MsgIBCTransfer) Route() string { return "ibc" }
func (msg MsgIBCTransfer) Type() string  { return "transfer" }

// x/bank/tx.go MsgSend.GetSigners()
func (msg MsgIBCTransfer) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{msg.SrcAddr} }

// get the sign bytes for ibc transfer message
func (msg MsgIBCTransfer) GetSignBytes() []byte {
	return msg.IBCPacket.GetSignBytes()
}

// validate ibc transfer message
func (msg MsgIBCTransfer) ValidateBasic() sdk.Error {
	return msg.IBCPacket.ValidateBasic()
}

// ----------------------------------
// MsgIBCReceive

// nolint - TODO rename to ReceiveMsg as folks will reference with ibc.ReceiveMsg
// MsgIBCReceive defines the message that a relayer uses to post an IBCPacket
// to the destination chain.
type MsgIBCReceive struct {
	IBCPacket
	Relayer  sdk.AccAddress
	Sequence uint64
}

// nolint
func (msg MsgIBCReceive) Route() string            { return "ibc" }
func (msg MsgIBCReceive) Type() string             { return "receive" }
func (msg MsgIBCReceive) ValidateBasic() sdk.Error { return msg.IBCPacket.ValidateBasic() }

// x/bank/tx.go MsgSend.GetSigners()
func (msg MsgIBCReceive) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{msg.Relayer} }

// get the sign bytes for ibc receive message
func (msg MsgIBCReceive) GetSignBytes() []byte {
	b, err := msgCdc.MarshalJSON(struct {
		IBCPacket json.RawMessage
		Relayer   sdk.AccAddress
		Sequence  uint64
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
