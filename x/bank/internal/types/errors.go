package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/bank module sentinel errors
var (
	ErrNoInputs            = sdkerrors.Register(ModuleName, 1, "no inputs to send transaction")
	ErrNoOutputs           = sdkerrors.Register(ModuleName, 2, "no outputs to send transaction")
	ErrInputOutputMismatch = sdkerrors.Register(ModuleName, 3, "sum inputs != sum outputs")
	ErrSendDisabled        = sdkerrors.Register(ModuleName, 4, "send transactions are disabled")
)
