package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/capability module sentinel errors
var (
	ErrCapabilityTaken          = sdkerrors.Register(ModuleName, 2, "capability name already taken")
	ErrOwnerClaimed             = sdkerrors.Register(ModuleName, 3, "given owner already claimed capability")
	ErrCapabilityNotOwned       = sdkerrors.Register(ModuleName, 4, "capability not owned by module")
	ErrCapabilityNotFound       = sdkerrors.Register(ModuleName, 5, "capability not found")
	ErrCapabilityOwnersNotFound = sdkerrors.Register(ModuleName, 6, "owners not found for capability")
)
