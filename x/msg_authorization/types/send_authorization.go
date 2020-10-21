package types

import (
	"reflect"

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

func (authorization SendAuthorization) MethodName() string {
	return "/cosmos.bank.v1beta1.Msg/Send"
}

func (authorization SendAuthorization) Accept(msg sdk.ServiceMsg, block tmproto.Header) (allow bool, updated Authorization, delete bool) {
	if reflect.TypeOf(msg.Request) == reflect.TypeOf(&bank.MsgSend{}) {
		msg := msg.Request.(*bank.MsgSend)
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
