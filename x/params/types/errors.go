package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/params module sentinel errors
var (
	ErrUnknownSubspace  = sdkerrors.Register(ModuleName, 1, "unknown subspace")
	ErrSettingParameter = sdkerrors.Register(ModuleName, 2, "failed to set parameter")
	ErrEmptyChanges     = sdkerrors.Register(ModuleName, 3, "submitted parameter changes are empty")
	ErrEmptySubspace    = sdkerrors.Register(ModuleName, 4, "parameter subspace is empty")
	ErrEmptyKey         = sdkerrors.Register(ModuleName, 5, "parameter key is empty")
	ErrEmptyValue       = sdkerrors.Register(ModuleName, 6, "parameter value is empty")
)
