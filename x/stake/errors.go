// nolint
package stake

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CodeType = sdk.CodeType

const (
	DefaultCodespace sdk.CodespaceType = 4

	// Gaia errors reserve 200 ~ 299.
	CodeInvalidValidator CodeType = 201
	CodeInvalidBond      CodeType = 202
	CodeInvalidInput     CodeType = 203
	CodeValidatorJailed  CodeType = 204
	CodeUnauthorized     CodeType = sdk.CodeUnauthorized
	CodeInternal         CodeType = sdk.CodeInternal
	CodeUnknownRequest   CodeType = sdk.CodeUnknownRequest
)

// NOTE: Don't stringer this, we'll put better messages in later.
func codeToDefaultMsg(code CodeType) string {
	switch code {
	case CodeInvalidValidator:
		return "Invalid Validator"
	case CodeInvalidBond:
		return "Invalid Bond"
	case CodeInvalidInput:
		return "Invalid Input"
	case CodeUnauthorized:
		return "Unauthorized"
	case CodeInternal:
		return "Internal Error"
	case CodeUnknownRequest:
		return "Unknown request"
	default:
		return sdk.CodeToDefaultMsg(code)
	}
}

//----------------------------------------
// Error constructors

func ErrNotEnoughBondShares(codespace sdk.CodespaceType, shares string) sdk.Error {
	return newError(codespace, CodeInvalidBond, fmt.Sprintf("not enough shares only have %v", shares))
}
func ErrValidatorEmpty(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeInvalidValidator, "Cannot bond to an empty validator")
}
func ErrBadBondingDenom(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeInvalidBond, "Invalid coin denomination")
}
func ErrBadBondingAmount(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeInvalidBond, "Amount must be > 0")
}
func ErrNoBondingAcct(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeInvalidValidator, "No bond account for this (address, validator) pair")
}
func ErrCommissionNegative(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeInvalidValidator, "Commission must be positive")
}
func ErrCommissionHuge(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeInvalidValidator, "Commission cannot be more than 100%")
}
func ErrBadValidatorAddr(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeInvalidValidator, "Validator does not exist for that address")
}
func ErrBadDelegatorAddr(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeInvalidValidator, "Delegator does not exist for that address")
}
func ErrValidatorExistsAddr(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeInvalidValidator, "Validator already exist, cannot re-declare candidacy")
}
func ErrValidatorRevoked(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeInvalidValidator, "Candidacy for this address is currently revoked")
}
func ErrMissingSignature(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeInvalidValidator, "Missing signature")
}
func ErrBondNotNominated(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeInvalidValidator, "Cannot bond to non-nominated account")
}
func ErrNoValidatorForAddress(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeInvalidValidator, "Validator does not exist for that address")
}
func ErrNoDelegatorForAddress(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeInvalidValidator, "Delegator does not contain validator bond")
}
func ErrInsufficientFunds(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeInvalidInput, "Insufficient bond shares")
}
func ErrBadShares(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeInvalidInput, "bad shares provided as input, must be MAX or decimal")
}
func ErrBadRemoveValidator(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeInvalidValidator, "Error removing validator")
}

//----------------------------------------

// TODO group with code from x/bank/errors.go

func msgOrDefaultMsg(msg string, code CodeType) string {
	if msg != "" {
		return msg
	}
	return codeToDefaultMsg(code)
}

func newError(codespace sdk.CodespaceType, code CodeType, msg string) sdk.Error {
	msg = msgOrDefaultMsg(msg, code)
	return sdk.NewError(codespace, code, msg)
}
