package types

import (
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ Authorization = &GenericAuthorization{}
)

func NewGenericAuthorization(methodName string) *GenericAuthorization {
	return &GenericAuthorization{
		MessageName: methodName,
	}
}

func (cap GenericAuthorization) MethodName() string {
	return cap.MessageName
}

func (cap GenericAuthorization) Accept(msg sdk.ServiceMsg, block tmproto.Header) (allow bool, updated Authorization, delete bool) {
	return true, &cap, false
}
