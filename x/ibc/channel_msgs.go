package ibc

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ----------------------------------------
// Msg Definitions

// MsgSend is a simpler form of sending ibc payload to another chain
// when there is no need for preparation
type MsgSend struct {
	Payload
	DestChain string
}

func (msg MsgSend) GetSignBytes() []byte {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

// MsgReceive defines the message that a relayer uses to post a packet
// to the destination chain.
type MsgReceive struct {
	Datagram
	Proof
	Relayer sdk.AccAddress
}

func (msg MsgReceive) GetSignBytes() []byte {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg MsgReceive) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Relayer}
}

type MsgCleanup struct {
	Sequence int64
	SrcChain string
	Proof    Proof
	Cleaner  sdk.AccAddress
}

func (msg MsgCleanup) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Cleaner}
}

func (msg MsgCleanup) Type() string {
	return "ibc"
}

func (msg MsgCleanup) ValidateBasic() sdk.Error {
	return nil
}
