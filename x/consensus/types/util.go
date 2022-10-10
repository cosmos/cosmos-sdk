package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Sentinel errors for the x/consensus module.
var (
	ErrUnauthorized = sdkerrors.Register(ModuleName, 2, "unauthorized action")
)

// Events
const (
	AttributeValueCategory = ModuleName

	EventTypeUpdateParam = "update_param"

	AttributeKeyParamUpdater = "param_updater"
)
