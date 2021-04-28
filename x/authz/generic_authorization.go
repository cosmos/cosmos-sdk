package authz

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

var (
	_ Authorization = &GenericAuthorization{}
)

// NewGenericAuthorization creates a new GenericAuthorization object.
func NewGenericAuthorization(methodName string) *GenericAuthorization {
	return &GenericAuthorization{
		Msg: methodName,
	}
}

// MsgTypeURL implements Authorization.MsgTypeURL.
func (a GenericAuthorization) MsgTypeURL() string {
	return a.Msg
}

// Accept implements Authorization.Accept.
func (a GenericAuthorization) Accept(ctx sdk.Context, msg sdk.ServiceMsg) (AcceptResponse, error) {
	return AcceptResponse{Accept: true}, nil
}

// ValidateBasic implements Authorization.ValidateBasic.
func (a GenericAuthorization) ValidateBasic() error {
	if !msgservice.IsServiceMsg(a.Msg) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidType, " %s is not a valid service msg", a.Msg)
	}
	return nil
}
