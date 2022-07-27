package authz

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_         Authorization = &CountAuthorization{}
	errMsgGt0               = "allowed authorizations must be greater than 0"
)

// NewCountAuthorization creates a new CountAuthorization object.
func NewCountAuthorization(msgTypeURL string, allowedAuthorizations int32) *CountAuthorization {
	return &CountAuthorization{
		Msg:                   msgTypeURL,
		AllowedAuthorizations: allowedAuthorizations,
	}
}

// MsgTypeURL implements Authorization.MsgTypeURL.
func (a CountAuthorization) MsgTypeURL() string {
	return a.Msg
}

// Accept implements Authorization.Accept.
func (a CountAuthorization) Accept(ctx sdk.Context, msg sdk.Msg) (AcceptResponse, error) {
	remaining, isNegative := a.decrement()
	if isNegative {
		return AcceptResponse{}, sdkerrors.ErrUnauthorized.Wrapf(errMsgGt0)
	}
	if remaining == 0 {
		return AcceptResponse{Accept: true, Delete: true}, nil
	}

	return AcceptResponse{Accept: true, Delete: false, Updated: &CountAuthorization{Msg: a.Msg, AllowedAuthorizations: remaining}}, nil
}

// ValidateBasic implements Authorization.ValidateBasic.
func (a CountAuthorization) ValidateBasic() error {
	if a.AllowedAuthorizations <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf(errMsgGt0)
	}
	return nil
}

func (a CountAuthorization) decrement() (int32, bool) {
	cnt := a.AllowedAuthorizations - 1
	return cnt, cnt < 0
}
