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
)

// ErrNoEvidenceHandlerExists returns a typed error for an invalid evidence
// handler route.
func ErrNoEvidenceHandlerExists(codespace sdk.CodespaceType, route string) error {
	return sdkerrors.New(
		string(codespace),
		uint32(CodeNoEvidenceHandlerExists),
		fmt.Sprintf("route '%s' does not have a registered evidence handler", route),
	)
}
