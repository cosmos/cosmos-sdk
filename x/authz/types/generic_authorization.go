package types

import (
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ Authorization = &GenericAuthorization{}
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
func (cap GenericAuthorization) Accept(msg sdk.ServiceMsg, block tmproto.Header) (updated Authorization, delete bool, err error) {
	return &cap, false, nil
}
