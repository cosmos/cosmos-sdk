package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// common IBC error codes
const (
	// DefaultCodespace of the IBC module
	DefaultCodespace sdk.CodespaceType = ModuleName

	CodeInvalidVersion sdk.CodeType = 101
)

// ErrInvalidVersion implements sdk.Error
func ErrInvalidVersion(codespace sdk.CodespaceType, msg string) error {
	return sdkerrors.Register(
		string(codespace),
		uint32(CodeInvalidVersion),
		msg,
	)
}
