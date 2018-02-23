// nolint
package stake

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CodeType = sdk.CodeType

const (
	// Gaia errors reserve 200 ~ 299.
	CodeInvalidValidator CodeType = 201
	CodeInvalidCandidate CodeType = 202
	CodeInvalidBond      CodeType = 203
	CodeInvalidInput     CodeType = 204
	CodeUnauthorized     CodeType = sdk.CodeUnauthorized
	CodeInternal         CodeType = sdk.CodeInternal
	CodeUnknownRequest   CodeType = sdk.CodeUnknownRequest
)

// NOTE: Don't stringer this, we'll put better messages in later.
func codeToDefaultMsg(code CodeType) string {
	switch code {
	case CodeInvalidValidator:
		return "Invalid Validator"
	case CodeInvalidCandidate:
		return "Invalid Candidate"
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

func ErrCandidateEmpty() error {
	return newError(CodeInvalidValidator, "Cannot bond to an empty candidate")
}
func ErrBadBondingDenom() error {
	return newError(CodeInvalidValidator, "Invalid coin denomination")
}
func ErrBadBondingAmount() error {
	return newError(CodeInvalidValidator, "Amount must be > 0")
}
func ErrNoBondingAcct() error {
	return newError(CodeInvalidValidator, "No bond account for this (address, validator) pair")
}
func ErrCommissionNegative() error {
	return newError(CodeInvalidValidator, "Commission must be positive")
}
func ErrCommissionHuge() error {
	return newError(CodeInvalidValidator, "Commission cannot be more than 100%")
}
func ErrBadValidatorAddr() error {
	return newError(CodeInvalidValidator, "Validator does not exist for that address")
}
func ErrCandidateExistsAddr() error {
	return newError(CodeInvalidValidator, "Candidate already exist, cannot re-declare candidacy")
}
func ErrMissingSignature() error {
	return newError(CodeInvalidValidator, "Missing signature")
}
func ErrBondNotNominated() error {
	return newError(CodeInvalidValidator, "Cannot bond to non-nominated account")
}
func ErrNoCandidateForAddress() error {
	return newError(CodeInvalidValidator, "Validator does not exist for that address")
}
func ErrNoDelegatorForAddress() error {
	return newError(CodeInvalidValidator, "Delegator does not contain validator bond")
}
func ErrInsufficientFunds() error {
	return newError(CodeInvalidValidator, "Insufficient bond shares")
}
func ErrBadRemoveValidator() error {
	return newError(CodeInvalidValidator, "Error removing validator")
}

//----------------------------------------

// TODO group with code from x/bank/errors.go

func msgOrDefaultMsg(msg string, code CodeType) string {
	if msg != "" {
		return msg
	}
	return codeToDefaultMsg(code)
}

func newError(code CodeType, msg string) sdk.Error {
	msg = msgOrDefaultMsg(msg, code)
	return sdk.NewError(code, msg)
}
