package ibc

import (
	"encoding/json"

	"github.com/tendermint/tendermint/lite"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgOpenConnection defines the message that is used for open a c
// that receives msg from another chain
type MsgOpenConnection struct {
	ROT      lite.FullCommit
	SrcChain string
	Signer   sdk.AccAddress
}

func (msg MsgOpenConnection) Type() string {
	return "ibc"
}

func (msg MsgOpenConnection) GetSignBytes() []byte {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg MsgOpenConnection) ValidateBasic() sdk.Error {
	if msg.ROT.Height() < 0 {
		// XXX: Codespace will be removed
		return ErrInvalidHeight(111)
	}
	return nil
}

func (msg MsgOpenConnection) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

type MsgUpdateConnection struct {
	SrcChain string
	Commit   lite.FullCommit
	//PacketProof
	Signer sdk.AccAddress
}

func (msg MsgUpdateConnection) Type() string {
	return "ibc"
}

func (msg MsgUpdateConnection) GetSignBytes() []byte {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg MsgUpdateConnection) ValidateBasic() sdk.Error {
	if msg.Commit.Commit.Height() < 0 {
		// XXX: Codespace will be removed
		return ErrInvalidHeight(111)
	}
	return nil
}

func (msg MsgUpdateConnection) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
