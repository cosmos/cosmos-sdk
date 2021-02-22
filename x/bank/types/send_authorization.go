package types

import (
	"reflect"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authz "github.com/cosmos/cosmos-sdk/x/authz/exported"
)

var (
	_ authz.Authorization = &SendAuthorization{}
)

// NewSendAuthorization creates a new SendAuthorization object.
func NewSendAuthorization(spendLimit sdk.Coins) *SendAuthorization {
	return &SendAuthorization{
		SpendLimit: spendLimit,
	}
}

// MethodName implements Authorization.MethodName.
func (authorization SendAuthorization) MethodName() string {
	return "/cosmos.bank.v1beta1.Msg/Send"
}

// Accept implements Authorization.Accept.
func (authorization SendAuthorization) Accept(msg sdk.ServiceMsg, block tmproto.Header) (updated authz.Authorization, delete bool, err error) {
	if reflect.TypeOf(msg.Request) == reflect.TypeOf(&MsgSend{}) {
		msg, ok := msg.Request.(*MsgSend)
		if ok {
			limitLeft, isNegative := authorization.SpendLimit.SafeSub(msg.Amount)
			if isNegative {
				return nil, false, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "requested amount is more than spend limit")
			}
			if limitLeft.IsZero() {
				return nil, true, nil
			}

			return &SendAuthorization{SpendLimit: limitLeft}, false, nil
		}
	}
	return nil, false, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "type mismatch")
}
