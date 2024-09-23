package types

import (
	coretransaction "cosmossdk.io/core/transaction"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ coretransaction.Msg = &MsgSend{}

// NewMsgSend constructs a msg to send coins from one account to another.
func NewMsgSend(fromAddr, toAddr string, amount sdk.Coins) *MsgSend {
	return &MsgSend{FromAddress: fromAddr, ToAddress: toAddr, Amount: amount}
}
