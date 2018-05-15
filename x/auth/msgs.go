package auth

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/tendermint/go-crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgClaimAccount - high level transaction of the auth module
type MsgClaimAccount struct {
	Address sdk.Address   `json:"address"`
	PubKey  crypto.PubKey `json:"public_key"`
}

var _ sdk.Msg = MsgClaimAccount{}

// NewMsgClaimAccount - msg to claim an account and set the PubKey
func NewMsgClaimAccount(addr sdk.Address, pubkey crypto.PubKey) MsgClaimAccount {
	return MsgClaimAccount{Address: addr, PubKey: pubkey}
}

// Implements Msg.
func (msg MsgClaimAccount) Type() string { return "auth" }

// Implements Msg.
func (msg MsgClaimAccount) ValidateBasic() sdk.Error {
	if bytes.Equal(msg.PubKey.Address(), msg.Address) {
		return sdk.ErrInvalidPubKey(fmt.Sprintf("PubKey is invalid"))
	}
	return nil
}

// Implements Msg.
func (msg MsgClaimAccount) Get(key interface{}) (value interface{}) {
	return nil
}

// Implements Msg.
func (msg MsgClaimAccount) GetSignBytes() []byte {
	b, err := json.Marshal(msg) // XXX: ensure some canonical form
	if err != nil {
		panic(err)
	}
	return b
}

// Implements Msg.
func (msg MsgClaimAccount) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Address}
}

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
	if len(msg.NewPubKey.Bytes()) == 0 {
		return sdk.ErrInvalidPubKey(fmt.Sprintf("New PubKey is invalid"))
	}
	return nil
}

// Implements Msg.
func (msg MsgChangeKey) Get(key interface{}) (value interface{}) {
	return nil
}

// Implements Msg.
func (msg MsgChangeKey) GetSignBytes() []byte {
	b, err := json.Marshal(msg) // XXX: ensure some canonical form
	if err != nil {
		panic(err)
	}
	return b
}

// Implements Msg.
func (msg MsgChangeKey) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Address}
}
