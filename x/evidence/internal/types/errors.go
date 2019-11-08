// DONTCOVER
package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Error codes specific to the evidence module
const (
	DefaultCodespace sdk.CodespaceType = ModuleName

	CodeNoEvidenceHandlerExists sdk.CodeType = 1
	CodeInvalidEvidence         sdk.CodeType = 2
	CodeNoEvidenceExists        sdk.CodeType = 3
	CodeEvidenceExists          sdk.CodeType = 4
)

// ErrNoEvidenceHandlerExists returns a typed ABCI error for an invalid evidence
// handler route.
func ErrNoEvidenceHandlerExists(codespace sdk.CodespaceType, route string) error {
	return sdkerrors.New(
		string(codespace),
		uint32(CodeNoEvidenceHandlerExists),
		fmt.Sprintf("route '%s' does not have a registered evidence handler", route),
	)
}

// ErrInvalidEvidence returns a typed ABCI error for invalid evidence.
func ErrInvalidEvidence(codespace sdk.CodespaceType, msg string) error {
	return sdkerrors.New(
		string(codespace),
		uint32(CodeInvalidEvidence),
		fmt.Sprintf("invalid evidence: %s", msg),
	)
}

// ErrNoEvidenceExists returns a typed ABCI error for Evidence that does not exist
// for a given hash.
func ErrNoEvidenceExists(codespace sdk.CodespaceType, hash string) error {
	return sdkerrors.New(
		string(codespace),
		uint32(CodeNoEvidenceExists),
		fmt.Sprintf("evidence with hash %s does not exist", hash),
	)
}

// ErrEvidenceExists returns a typed ABCI error for Evidence that already exists
// by hash in state.
func ErrEvidenceExists(codespace sdk.CodespaceType, hash string) error {
	return sdkerrors.New(
		string(codespace),
		uint32(CodeEvidenceExists),
		fmt.Sprintf("evidence with hash %s already exists", hash),
	)
}
