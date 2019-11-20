package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CodeType = sdk.CodeType

const (
	// Default slashing codespace
	DefaultCodespace sdk.CodespaceType = ModuleName

	CodeInvalidGranter        CodeType = 101
	CodeInvalidGrantee        CodeType = 102
	CodeInvalidExpirationTime CodeType = 103
)

func ErrInvalidGranter(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidGranter, "invalid granter address")
}

func ErrInvalidGrantee(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidGrantee, "invalid grantee address")
}
func ErrInvalidExpirationTime(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidExpirationTime, "expiration time of authorization should be more than current time")
}
