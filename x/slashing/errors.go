//nolint
package slashing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Local code type
type CodeType = sdk.CodeType

const (
	// Default slashing codespace
	DefaultCodespace sdk.CodespaceType = 10

	// Invalid validator
	CodeInvalidValidator CodeType = 201
	// Validator jailed
	CodeValidatorJailed CodeType = 202
)

func ErrNoValidatorForAddress(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeInvalidValidator, "That address is not associated with any known validator")
}
func ErrBadValidatorAddr(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeInvalidValidator, "Validator does not exist for that address")
}
func ErrValidatorJailed(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeValidatorJailed, "Validator jailed, cannot yet be unrevoked")
}

func codeToDefaultMsg(code CodeType) string {
	switch code {
	case CodeInvalidValidator:
		return "Invalid Validator"
	case CodeValidatorJailed:
		return "Validator Jailed"
	default:
		return sdk.CodeToDefaultMsg(code)
	}
}

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
