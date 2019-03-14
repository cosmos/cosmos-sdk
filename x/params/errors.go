package params

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// nolint
const (
	DefaultCodespace sdk.CodespaceType = "params"

	CodeUnknownSubspace  sdk.CodeType = 1
	CodeSettingParameter sdk.CodeType = 2
	CodeEmptyData        sdk.CodeType = 3
)

// nolint
func ErrUnknownSubspace(codespace sdk.CodespaceType, space string) sdk.Error {
	return sdk.NewError(codespace, CodeUnknownSubspace, fmt.Sprintf("Unknown subspace %s", space))
}
func ErrSettingParameter(codespace sdk.CodespaceType, key, subkey, value []byte, msg string) sdk.Error {
	return sdk.NewError(codespace, CodeSettingParameter, fmt.Sprintf("Error setting parameter %X on %s (%X); msg: \"%s\"", value, string(key), subkey, msg))
}
func ErrEmptyChanges(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeEmptyData, "Submitted parameter changes are empty")
}
func ErrEmptySpace(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeEmptyData, "Parameter space is empty")
}
func ErrEmptyKey(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeEmptyData, "Parameter key is empty")
}
func ErrEmptyValue(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeEmptyData, "Parameter value is empty")
}
