// nolint
package supply

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CodeType = sdk.CodeType

// define the supply errors codes
const (
	DefaultCodespace         sdk.CodespaceType = ModuleName
	CodeInsufficientSupply   CodeType          = 101
	CodeInsufficientHoldings CodeType          = 102
)

func ErrInsufficientSupply(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(
		codespace,
		CodeInsufficientSupply,
		"requested tokens greater than current total supply",
	)
}

func ErrInsufficientHoldings(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(
		codespace,
		CodeInsufficientHoldings,
		"insufficient token holdings",
	)
}
