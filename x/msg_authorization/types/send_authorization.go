package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

var (
	_ Authorization = &SendAuthorization{}
)

func NewSendAuthorization(spendLimit sdk.Coins) *SendAuthorization {
	return &SendAuthorization{
		SpendLimit: spendLimit,
	}
}

func (authorization SendAuthorization) MsgType() string {
	return bank.MsgSend{}.Type()
}

func (authorization SendAuthorization) Accept(msg sdk.Msg, block tmproto.Header) (allow bool, updated Authorization, delete bool) {
	switch msg := msg.(type) {
	case *bank.MsgSend:
		limitLeft, isNegative := authorization.SpendLimit.SafeSub(msg.Amount)
		if isNegative {
			return false, nil, false
		}
		if limitLeft.IsZero() {
			return true, nil, true
		}

		return true, &SendAuthorization{SpendLimit: limitLeft}, false
	}
	return false, nil, false
}
