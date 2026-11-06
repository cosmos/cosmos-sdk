package types

import errorsmod "cosmossdk.io/errors"

// x/safety module sentinel errors. Invalid addresses and unauthorized senders
// use sdkerrors.ErrInvalidAddress and sdkerrors.ErrUnauthorized, as the
// authority check in sdk.ValidateAuthority does.
var (
	ErrInvalidPattern   = errorsmod.Register(ModuleName, 2, "invalid pattern")
	ErrEmptyPatterns    = errorsmod.Register(ModuleName, 3, "pattern list is empty")
	ErrDuplicatePattern = errorsmod.Register(ModuleName, 4, "duplicate pattern")
	ErrInvalidParams    = errorsmod.Register(ModuleName, 5, "invalid params")
	ErrOutOfScope       = errorsmod.Register(ModuleName, 6, "pattern not covered by operator scope")
	ErrNoMatchingMsg    = errorsmod.Register(ModuleName, 7, "pattern matches no registered unprotected msg type")
	ErrAlreadyDisabled  = errorsmod.Register(ModuleName, 8, "pattern already disabled")
	ErrNotDisabled      = errorsmod.Register(ModuleName, 9, "pattern not in disabled list")
	ErrMsgDisabled      = errorsmod.Register(ModuleName, 10, "msg type disabled by x/safety")
)
