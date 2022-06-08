package errors

import sdkerrors "cosmossdk.io/errors"

// mathCodespace is the codespace for all errors defined in math package
const mathCodespace = "math"

// ErrInvalidDecString defines an error for an invalid decimal string
var ErrInvalidDecString = sdkerrors.Register(mathCodespace, 10, "invalid decimal string")
