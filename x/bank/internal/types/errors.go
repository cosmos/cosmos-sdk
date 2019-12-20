package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/bank module errors that reserve codes 100-199
var (
	ErrNoInputs            = sdkerrors.Register(ModuleName, 100, "no inputs to send transaction")
	ErrNoOutputs           = sdkerrors.Register(ModuleName, 101, "no outputs to send transaction")
	ErrInputOutputMismatch = sdkerrors.Register(ModuleName, 102, "sum inputs != sum outputs")
	ErrSendDisabled        = sdkerrors.Register(ModuleName, 103, "send transactions are disabled")
)
