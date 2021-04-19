package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authz "github.com/cosmos/cosmos-sdk/x/authz/exported"
)

var (
	_ authz.Authorization = &GenericAuthorization{}
)

// NewGenericAuthorization creates a new GenericAuthorization object.
func NewGenericAuthorization(methodName string) *GenericAuthorization {
	return &GenericAuthorization{
		MsgTypeUrl: methodName,
	}
}

// MethodName implements Authorization.MethodName.
func (a GenericAuthorization) MethodName() string {
	return a.MsgTypeUrl
}

// Accept implements Authorization.Accept.
func (a GenericAuthorization) Accept(ctx sdk.Context, msg sdk.ServiceMsg) (authz.AcceptResponse, error) {
	return authz.AcceptResponse{Accept: true}, nil
}

// ValidateBasic implements Authorization.ValidateBasic.
func (a GenericAuthorization) ValidateBasic() error {
	if !msgservice.IsServiceMsg(a.MsgTypeUrl) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidType, " %s is not a valid service msg", a.MsgTypeUrl)
	}
	return nil
}
