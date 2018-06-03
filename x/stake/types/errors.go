// nolint
package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CodeType = sdk.CodeType

const (
	DefaultCodespace sdk.CodespaceType = 4

	CodeInvalidValidator CodeType = 101
	CodeInvalidBond      CodeType = 102
	CodeInvalidInput     CodeType = 103
	CodeValidatorJailed  CodeType = 104
	CodeUnauthorized     CodeType = sdk.CodeUnauthorized
	CodeInternal         CodeType = sdk.CodeInternal
	CodeUnknownRequest   CodeType = sdk.CodeUnknownRequest
)

func ErrNotEnoughBondShares(codespace sdk.CodespaceType, shares string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidBond, fmt.Sprintf("not enough shares only have %v", shares))
}
func ErrValidatorEmpty(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidValidator, "Cannot bond to an empty validator")
}
func ErrBadBondingDenom(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidBond, "Invalid coin denomination")
}
func ErrBadBondingAmount(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidBond, "Amount must be > 0")
}
func ErrBadSharesAmount(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidBond, "Shares must be > 0")
}
func ErrBadSharesPercent(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidBond, "Shares percent must be >0 and <=1")
}
func ErrNoBondingAcct(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidValidator, "No bond account for this (address, validator) pair")
}
func ErrCommissionNegative(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidValidator, "Commission must be positive")
}
func ErrCommissionHuge(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidValidator, "Commission cannot be more than 100%")
}
func ErrBadValidatorAddr(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidValidator, "Validator does not exist for that address")
}
func ErrBothShareMsgsGiven(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInput, "Both shares amount and shares percent provided")
}
func ErrNeitherShareMsgsGiven(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInput, "Neither shares amount nor shares percent provided")
}
func ErrBadDelegatorAddr(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidValidator, "Delegator does not exist for that address")
}
func ErrValidatorExistsAddr(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidValidator, "Validator already exist, cannot re-create validator")
}
func ErrValidatorRevoked(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidValidator, "Validator for this address is currently revoked")
}
func ErrMissingSignature(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidValidator, "Missing signature")
}
func ErrBondNotNominated(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidValidator, "Cannot bond to non-nominated account")
}
func ErrNoValidatorForAddress(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidValidator, "Validator does not exist for that address")
}
func ErrNoDelegatorForAddress(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidValidator, "Delegator does not contain validator bond")
}
func ErrInsufficientFunds(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInput, "Insufficient bond shares")
}
func ErrBadRemoveValidator(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidValidator, "Error removing validator")
}
