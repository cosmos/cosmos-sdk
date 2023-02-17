package proposal

import (
	errorsmod "cosmossdk.io/errors"
)

// x/params module sentinel errors
var (
	ErrUnknownSubspace  = errorsmod.Register(ModuleName, 2, "unknown subspace")
	ErrSettingParameter = errorsmod.Register(ModuleName, 3, "failed to set parameter")
	ErrEmptyChanges     = errorsmod.Register(ModuleName, 4, "submitted parameter changes are empty")
	ErrEmptySubspace    = errorsmod.Register(ModuleName, 5, "parameter subspace is empty")
	ErrEmptyKey         = errorsmod.Register(ModuleName, 6, "parameter key is empty")
	ErrEmptyValue       = errorsmod.Register(ModuleName, 7, "parameter value is empty")
)
