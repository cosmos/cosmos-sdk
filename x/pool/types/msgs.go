package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg = &MsgFundCommunityPool{}
	_ sdk.Msg = &MsgCommunityPoolSpend{}
)

// NewMsgFundCommunityPool creates a new MsgFundCommunityPool instance.
func NewMsgFundCommunityPool(amount sdk.Coins, depositor string) *MsgFundCommunityPool {
	return &MsgFundCommunityPool{
		Amount:    amount,
		Depositor: depositor,
	}
}

// NewMsgCommunityPoolSpend creates a new MsgCommunityPoolSpend instance.
func NewMsgCommunityPoolSpend(authority, recipient string, amount sdk.Coins) *MsgCommunityPoolSpend {
	return &MsgCommunityPoolSpend{
		Authority: authority,
		Recipient: recipient,
		Amount:    amount,
	}
}
