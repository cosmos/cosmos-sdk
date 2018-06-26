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
	return newError(codespace, CodeInvalidValidator, "that address is not associated with any known validator")
}
func ErrBadValidatorAddr(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeInvalidValidator, "validator does not exist for that address")
}
func ErrValidatorJailed(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeValidatorJailed, "validator jailed, cannot yet be unrevoked")
}

func codeToDefaultMsg(code CodeType) string {
	switch code {
	case CodeInvalidValidator:
		return "invalid Validator"
	case CodeValidatorJailed:
		return "validator Jailed"
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
