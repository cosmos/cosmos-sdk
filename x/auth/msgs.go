package auth

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	crypto "github.com/tendermint/go-crypto"
)

// MsgChangeKey - high level transaction of the auth module
type MsgChangeKey struct {
	Address   sdk.Address   `json:"address"`
	NewPubKey crypto.PubKey `json:"public_key"`
}

var _ sdk.Msg = MsgChangeKey{}

// NewMsgChangeKey - msg to claim an account and set the PubKey
func NewMsgChangeKey(addr sdk.Address, pubkey crypto.PubKey) MsgChangeKey {
	return MsgChangeKey{Address: addr, NewPubKey: pubkey}
}

// Implements Msg.
func (msg MsgChangeKey) Type() string { return "auth" }

// Implements Msg.
func (msg MsgChangeKey) ValidateBasic() sdk.Error {
	return nil
}

// Implements Msg.
func (msg MsgChangeKey) GetSignBytes() []byte {
	b, err := msgCdc.MarshalJSON(msg) // XXX: ensure some canonical form
	if err != nil {
		panic(err)
	}
	return b
}

// Implements Msg.
func (msg MsgChangeKey) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Address}
}
