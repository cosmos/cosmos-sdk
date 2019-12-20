package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/params module errors that reserve codes 600-699
var (
	ErrUnknownSubspace  = sdkerrors.Register(ModuleName, 600, "unknown subspace")
	ErrSettingParameter = sdkerrors.Register(ModuleName, 601, "failed to set parameter")
	ErrEmptyChanges     = sdkerrors.Register(ModuleName, 602, "submitted parameter changes are empty")
	ErrEmptySubspace    = sdkerrors.Register(ModuleName, 603, "parameter subspace is empty")
	ErrEmptyKey         = sdkerrors.Register(ModuleName, 604, "parameter key is empty")
	ErrEmptyValue       = sdkerrors.Register(ModuleName, 605, "parameter value is empty")
)
