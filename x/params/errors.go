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
)

// nolint
func ErrUnknownSubspace(codespace sdk.CodespaceType, space string) sdk.Error {
	return sdk.NewError(codespace, CodeUnknownSubspace, fmt.Sprintf("Unknown subspace %s", space))
}
func ErrSettingParameter(codespace sdk.CodespaceType, key, subkey, value []byte, msg string) sdk.Error {
	return sdk.NewError(codespace, CodeSettingParameter, fmt.Sprintf("Error setting parameter %X on %s (%X); msg: \"%s\"", value, string(key), subkey, msg))
}
