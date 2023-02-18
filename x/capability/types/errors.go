package types

import "cosmossdk.io/errors"

// x/capability module sentinel errors
var (
	ErrInvalidCapabilityName    = errors.Register(ModuleName, 2, "capability name not valid")
	ErrNilCapability            = errors.Register(ModuleName, 3, "provided capability is nil")
	ErrCapabilityTaken          = errors.Register(ModuleName, 4, "capability name already taken")
	ErrOwnerClaimed             = errors.Register(ModuleName, 5, "given owner already claimed capability")
	ErrCapabilityNotOwned       = errors.Register(ModuleName, 6, "capability not owned by module")
	ErrCapabilityNotFound       = errors.Register(ModuleName, 7, "capability not found")
	ErrCapabilityOwnersNotFound = errors.Register(ModuleName, 8, "owners not found for capability")
)
