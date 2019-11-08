package host

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// SubModuleName defines the ICS 24 host
const SubModuleName = "host"

// Error codes specific to the ibc host submodule
const (
	DefaultCodespace sdk.CodespaceType = SubModuleName

	CodeInvalidID     sdk.CodeType = 1
	CodeInvalidPath   sdk.CodeType = 2
	CodeInvalidPacket sdk.CodeType = 3
)

// ErrInvalidID returns a typed ABCI error for an invalid identifier
func ErrInvalidID(codespace sdk.CodespaceType, ID string) error {
	return sdkerrors.New(
		string(codespace),
		uint32(CodeInvalidID),
		fmt.Sprintf("invalid identifier '%s", ID),
	)
}

// ErrInvalidPath returns a typed ABCI error for an invalid path
func ErrInvalidPath(codespace sdk.CodespaceType, path string) error {
	return sdkerrors.New(
		string(codespace),
		uint32(CodeInvalidPath),
		fmt.Sprintf("invalid path '%s", path),
	)
}

// ErrInvalidPacket returns a typed ABCI error for an invalid identifier
func ErrInvalidPacket(codespace sdk.CodespaceType, msg string) error {
	return sdkerrors.New(
		string(codespace),
		uint32(CodeInvalidPacket),
		fmt.Sprintf("invalid packet: %s", msg),
	)
}
