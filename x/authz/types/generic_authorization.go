package types

import (
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

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
func (cap GenericAuthorization) MethodName() string {
	return cap.MessageName
}

// Accept implements Authorization.Accept.
func (cap GenericAuthorization) Accept(msg sdk.ServiceMsg, block tmproto.Header) (updated exported.Authorization, delete bool, err error) {
	return &cap, false, nil
}
