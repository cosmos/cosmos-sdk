package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// common IBC error codes
const (
	// DefaultCodespace of the IBC module
	DefaultCodespace sdk.CodespaceType = ModuleName

	CodeInvalidProof   sdk.CodeType = 234
	CodeInvalidVersion sdk.CodeType = 235
)

// ErrInvalidProof implements sdk.Error
func ErrInvalidProof(codespace sdk.CodespaceType, msg string) error {
	return sdkerrors.New(
		string(codespace),
		uint32(CodeInvalidProof),
		msg,
	)
}

// ErrInvalidVersion implements sdk.Error
func ErrInvalidVersion(codespace sdk.CodespaceType, msg string) error {
	return sdkerrors.New(
		string(codespace),
		uint32(CodeInvalidVersion),
		msg,
	)
}
