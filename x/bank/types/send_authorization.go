package types

import (
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
	return sdk.MsgTypeURL(&MsgSend{})
}

// Accept implements Authorization.Accept.
func (authorization SendAuthorization) Accept(_ sdk.Context, msg sdk.Msg) (updated authz.Authorization, delete bool, err error) {
	msgSend, ok := msg.(*MsgSend)
	if !ok {
		return nil, false, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "type mismatch")
	}

	limitLeft, isNegative := authorization.SpendLimit.SafeSub(msgSend.Amount)
	if isNegative {
		return nil, false, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "requested amount is more than spend limit")
	}
	if limitLeft.IsZero() {
		return nil, true, nil
	}

	return &SendAuthorization{SpendLimit: limitLeft}, false, nil
}

// ValidateBasic implements Authorization.ValidateBasic.
func (authorization SendAuthorization) ValidateBasic() error {
	if authorization.SpendLimit == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "spend limit cannot be nil")
	}
	if !authorization.SpendLimit.IsAllPositive() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "spend limit cannot be negitive")
	}
	return nil
}
