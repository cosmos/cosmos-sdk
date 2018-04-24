package auth

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/tendermint/go-crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgSend - high level transaction of the coin module
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
	if bytes.Equal(msg.PubKey.Address(), msg.Address.Bytes()) {
		return sdk.ErrInvalidPubKey(fmt.Sprintf("PubKey does not match Signer address %v", msg.Address))
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
