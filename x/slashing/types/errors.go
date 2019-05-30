//nolint
package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Local code type
type CodeType = sdk.CodeType

const (
	// Default slashing codespace
	DefaultCodespace sdk.CodespaceType = ModuleName

	CodeInvalidValidator      CodeType = 101
	CodeValidatorJailed       CodeType = 102
	CodeValidatorNotJailed    CodeType = 103
	CodeMissingSelfDelegation CodeType = 104
	CodeSelfDelegationTooLow  CodeType = 105
	CodeMissingSigningInfo    CodeType = 106
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

func ErrMissingSelfDelegation(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeMissingSelfDelegation, "validator has no self-delegation; cannot be unjailed")
}

func ErrSelfDelegationTooLowToUnjail(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeValidatorNotJailed, "validator's self delegation less than MinSelfDelegation, cannot be unjailed")
}

func ErrNoSigningInfoFound(codespace sdk.CodespaceType, consAddr sdk.ConsAddress) sdk.Error {
	return sdk.NewError(codespace, CodeMissingSigningInfo, fmt.Sprintf("no signing info found for address: %s", consAddr))
}
