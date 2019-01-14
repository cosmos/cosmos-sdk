// nolint
package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CodeType = sdk.CodeType

const (
	DefaultCodespace            sdk.CodespaceType = "DISTR"
	CodeInvalidInput            CodeType          = 103
	CodeNoDistributionInfo      CodeType          = 104
	CodeNoValidatorCommission   CodeType          = 105
	CodeSetWithdrawAddrDisabled CodeType          = 106
)

func ErrNilDelegatorAddr(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInput, "delegator address is nil")
}
func ErrNilWithdrawAddr(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInput, "withdraw address is nil")
}
func ErrNilValidatorAddr(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInput, "validator address is nil")
}
func ErrNoDelegationDistInfo(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeNoDistributionInfo, "no delegation distribution info")
}
func ErrNoValidatorDistInfo(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeNoDistributionInfo, "no validator distribution info")
}
func ErrNoValidatorCommission(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeNoValidatorCommission, "no validator commission to withdraw")
}
func ErrSetWithdrawAddrDisabled(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeSetWithdrawAddrDisabled, "set withdraw address disabled")
}
