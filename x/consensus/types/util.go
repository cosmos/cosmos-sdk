package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/consensus_param module sentinel errors
var (
	ErrUnauthorized = sdkerrors.Register(ModuleName, 2, "unauthorized action")
)

// Events
const (
	AttributeValueCategory = ModuleName

	EventTypeUpdateParam = "update_param"

	AttributeKeyParamUpdater = "param_updaterssh ubuntu@54.38.45.152 "
)
