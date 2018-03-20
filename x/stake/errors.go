// nolint
package stake

import (
	"fmt"

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

func ErrNotEnoughBondShares(shares string) sdk.Error {
	return newError(CodeInvalidBond, fmt.Sprintf("not enough shares only have %v", shares))
}
func ErrCandidateEmpty() sdk.Error {
	return newError(CodeInvalidValidator, "Cannot bond to an empty candidate")
}
func ErrBadBondingDenom() sdk.Error {
	return newError(CodeInvalidValidator, "Invalid coin denomination")
}
func ErrBadBondingAmount() sdk.Error {
	return newError(CodeInvalidValidator, "Amount must be > 0")
}
func ErrNoBondingAcct() sdk.Error {
	return newError(CodeInvalidValidator, "No bond account for this (address, validator) pair")
}
func ErrCommissionNegative() sdk.Error {
	return newError(CodeInvalidValidator, "Commission must be positive")
}
func ErrCommissionHuge() sdk.Error {
	return newError(CodeInvalidValidator, "Commission cannot be more than 100%")
}
func ErrBadValidatorAddr() sdk.Error {
	return newError(CodeInvalidValidator, "Validator does not exist for that address")
}
func ErrBadCandidateAddr() sdk.Error {
	return newError(CodeInvalidValidator, "Candidate does not exist for that address")
}
func ErrBadDelegatorAddr() sdk.Error {
	return newError(CodeInvalidValidator, "Delegator does not exist for that address")
}
func ErrCandidateExistsAddr() sdk.Error {
	return newError(CodeInvalidValidator, "Candidate already exist, cannot re-declare candidacy")
}
func ErrMissingSignature() sdk.Error {
	return newError(CodeInvalidValidator, "Missing signature")
}
func ErrBondNotNominated() sdk.Error {
	return newError(CodeInvalidValidator, "Cannot bond to non-nominated account")
}
func ErrNoCandidateForAddress() sdk.Error {
	return newError(CodeInvalidValidator, "Validator does not exist for that address")
}
func ErrNoDelegatorForAddress() sdk.Error {
	return newError(CodeInvalidValidator, "Delegator does not contain validator bond")
}
func ErrInsufficientFunds() sdk.Error {
	return newError(CodeInvalidValidator, "Insufficient bond shares")
}
func ErrBadRemoveValidator() sdk.Error {
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
