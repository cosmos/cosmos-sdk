// nolint
package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CodeType = sdk.CodeType

const (
	DefaultCodespace sdk.CodespaceType = 4

	CodeInvalidValidator  CodeType = 101
	CodeInvalidDelegation CodeType = 102
	CodeInvalidInput      CodeType = 103
	CodeValidatorJailed   CodeType = 104
	CodeUnauthorized      CodeType = sdk.CodeUnauthorized
	CodeInternal          CodeType = sdk.CodeInternal
	CodeUnknownRequest    CodeType = sdk.CodeUnknownRequest
)

//validator
func ErrNilValidatorAddr(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInput, "validator address is nil")
}

func ErrNoValidatorFound(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidValidator, "validator does not exist for that address")
}

func ErrValidatorOwnerExists(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidValidator, "validator already exist for this owner-address, must use new validator-owner address")
}

func ErrValidatorPubKeyExists(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidValidator, "validator already exist for this pubkey, must use new validator pubkey")
}

func ErrValidatorRevoked(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidValidator, "validator for this address is currently revoked")
}

func ErrBadRemoveValidator(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidValidator, "error removing validator")
}

func ErrDescriptionLength(codespace sdk.CodespaceType, descriptor string, got, max int) sdk.Error {
	msg := fmt.Sprintf("bad description length for %v, got length %v, max is %v", descriptor, got, max)
	return sdk.NewError(codespace, CodeInvalidValidator, msg)
}

func ErrCommissionNegative(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidValidator, "commission must be positive")
}

func ErrCommissionHuge(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidValidator, "commission cannot be more than 100%")
}

func ErrNilDelegatorAddr(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInput, "delegator address is nil")
}

func ErrBadDenom(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidDelegation, "invalid coin denomination")
}

func ErrBadDelegationAmount(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidDelegation, "amount must be > 0")
}

func ErrNoDelegation(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidDelegation, "no delegation for this (address, validator) pair")
}

func ErrBadDelegatorAddr(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidDelegation, "delegator does not exist for that address")
}

func ErrNoDelegatorForAddress(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidDelegation, "delegator does not contain this delegation")
}

func ErrInsufficientShares(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidDelegation, "insufficient delegation shares")
}

func ErrDelegationValidatorEmpty(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidDelegation, "cannot delegate to an empty validator")
}

func ErrNotEnoughDelegationShares(codespace sdk.CodespaceType, shares string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidDelegation, fmt.Sprintf("not enough shares only have %v", shares))
}

func ErrBadSharesAmount(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidDelegation, "shares must be > 0")
}

func ErrBadSharesPrecision(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidDelegation,
		fmt.Sprintf("shares denominator must be < %s, try reducing the number of decimal points",
			maximumBondingRationalDenominator.String()),
	)
}

func ErrBadSharesPercent(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidDelegation, "shares percent must be >0 and <=1")
}

func ErrNotMature(codespace sdk.CodespaceType, operation, descriptor string, got, min int64) sdk.Error {
	msg := fmt.Sprintf("%v is not mature requires a min %v of %v, currently it is %v",
		operation, descriptor, got, min)
	return sdk.NewError(codespace, CodeUnauthorized, msg)
}

func ErrNoUnbondingDelegation(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidDelegation, "no unbonding delegation found")
}

func ErrNoRedelegation(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidDelegation, "no redelegation found")
}

func ErrBadRedelegationDst(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidDelegation, "redelegation validator not found")
}

func ErrTransitiveRedelegation(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidDelegation,
		"redelegation to this validator already in progress, first redelegation to this validator must complete before next redelegation")
}

func ErrBothShareMsgsGiven(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInput, "both shares amount and shares percent provided")
}

func ErrNeitherShareMsgsGiven(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInput, "neither shares amount nor shares percent provided")
}

func ErrMissingSignature(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidValidator, "missing signature")
}
