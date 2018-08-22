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

	CodeInvalidValidator   CodeType = 101
	CodeValidatorJailed    CodeType = 102
	CodeValidatorNotJailed CodeType = 103
)

func ErrNoValidatorForAddress(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidValidator, "that address is not associated with any known validator")
}
func ErrBadValidatorAddr(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidValidator, "validator does not exist for that address")
}
func ErrValidatorJailed(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeValidatorJailed, "validator still jailed, cannot yet be unjailed")
}
func ErrValidatorNotJailed(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeValidatorNotJailed, "validator not jailed, cannot be unjailed")
}
