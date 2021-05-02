package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz/exported"
)

var (
	_ exported.Authorization = &GenericAuthorization{}
)

// NewGenericAuthorization creates a new GenericAuthorization object.
func NewGenericAuthorization(methodName string) *GenericAuthorization {
	return &GenericAuthorization{
		MessageName: methodName,
	}
}

// MethodName implements Authorization.MethodName.
func (authorization GenericAuthorization) MethodName() string {
	return authorization.MessageName
}

// Accept implements Authorization.Accept.
func (authorization GenericAuthorization) Accept(ctx sdk.Context, msg sdk.Msg) (updated exported.Authorization, delete bool, err error) {
	return &authorization, false, nil
}

// ValidateBasic implements Authorization.ValidateBasic.
func (authorization GenericAuthorization) ValidateBasic() error {
	return nil
}
