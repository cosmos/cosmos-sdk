package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	SubModuleName = "solo machine"
)

var (
	ErrInvalidHeader               = sdkerrors.Register(SubModuleName, 2, "invalid header")
	ErrInvalidSequence             = sdkerrors.Register(SubModuleName, 3, "invalid sequence")
	ErrInvalidSignatureAndData     = sdkerrors.Register(SubModuleName, 4, "invalid signature and data")
	ErrSignatureVerificationFailed = sdkerrors.Register(SubModuleName, 5, "signature verification failed")
	ErrInvalidProof                = sdkerrors.Register(SubModuleName, 6, "invalid solo machine proof")
	ErrInvalidDataType             = sdkerrors.Register(SubModuleName, 7, "invalid data type")
)
