package authz

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/authz"
)

// NewGenericAuthorization creates a new GenericAuthorization object.
func NewGenericAuthorization(msgTypeURL string) *GenericAuthorization {
	return &GenericAuthorization{
		Msg: msgTypeURL,
	}
}

// MsgTypeURL implements Authorization.MsgTypeURL.
func (a GenericAuthorization) MsgTypeURL() string {
	return a.Msg
}

// Accept implements Authorization.Accept.
func (a GenericAuthorization) Accept(ctx context.Context, msg sdk.Msg) (authz.AcceptResponse, error) {
	return authz.AcceptResponse{Accept: true}, nil
}

// ValidateBasic implements Authorization.ValidateBasic.
func (a GenericAuthorization) ValidateBasic() error {
	if a.Msg == "" {
		return errors.New("msg type cannot be empty")
	}
	return nil
}

// Accept implements Authorization.Accept.
// func (a SendAuthorization) Accept(ctx context.Context, msg sdk.Msg) (authz.AcceptResponse, error) {
// 	mSend, ok := msg.(*MsgSend)
// 	if !ok {
// 		return authz.AcceptResponse{}, sdkerrors.ErrInvalidType.Wrap("type mismatch")
// 	}

// 	limitLeft, isNegative := a.SpendLimit.SafeSub(mSend.Amount...)
// 	if isNegative {
// 		return authz.AcceptResponse{}, sdkerrors.ErrInsufficientFunds.Wrapf("requested amount is more than spend limit")
// 	}

// 	authzEnv, ok := ctx.Value(corecontext.EnvironmentContextKey).(appmodule.Environment)
// 	if !ok {
// 		return authz.AcceptResponse{}, sdkerrors.ErrUnauthorized.Wrap("environment not set")
// 	}

// 	isAddrExists := false
// 	toAddr := mSend.ToAddress
// 	allowedList := a.GetAllowList()
// 	for _, addr := range allowedList {
// 		if err := authzEnv.GasService.GasMeter(ctx).Consume(gasCostPerIteration, "send authorization"); err != nil {
// 			return authz.AcceptResponse{}, err
// 		}

// 		if addr == toAddr {
// 			isAddrExists = true
// 			break
// 		}
// 	}

// 	if len(allowedList) > 0 && !isAddrExists {
// 		return authz.AcceptResponse{}, sdkerrors.ErrUnauthorized.Wrapf("cannot send to %s address", toAddr)
// 	}

// 	if limitLeft.IsZero() {
// 		return authz.AcceptResponse{Accept: true, Delete: true}, nil
// 	}

// 	return authz.AcceptResponse{Accept: true, Delete: false, Updated: &SendAuthorization{SpendLimit: limitLeft, AllowList: allowedList}}, nil
// }
