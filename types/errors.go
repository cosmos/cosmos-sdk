package types

import (
	tmsp "github.com/tendermint/tmsp/types"
)

var (
	ErrInternalError        = tmsp.NewError(tmsp.CodeType_InternalError, "Internal error")
	ErrDuplicateAddress     = tmsp.NewError(tmsp.CodeType_BaseDuplicateAddress, "Error duplicate address")
	ErrEncodingError        = tmsp.NewError(tmsp.CodeType_BaseEncodingError, "Error encoding error")
	ErrInsufficientFees     = tmsp.NewError(tmsp.CodeType_BaseInsufficientFees, "Error insufficient fees")
	ErrInsufficientFunds    = tmsp.NewError(tmsp.CodeType_BaseInsufficientFunds, "Error insufficient funds")
	ErrInsufficientGasPrice = tmsp.NewError(tmsp.CodeType_BaseInsufficientGasPrice, "Error insufficient gas price")
	ErrInvalidAddress       = tmsp.NewError(tmsp.CodeType_BaseInvalidAddress, "Error invalid address")
	ErrInvalidAmount        = tmsp.NewError(tmsp.CodeType_BaseInvalidAmount, "Error invalid amount")
	ErrInvalidPubKey        = tmsp.NewError(tmsp.CodeType_BaseInvalidPubKey, "Error invalid pubkey")
	ErrInvalidSequence      = tmsp.NewError(tmsp.CodeType_BaseInvalidSequence, "Error invalid sequence")
	ErrInvalidSignature     = tmsp.NewError(tmsp.CodeType_BaseInvalidSignature, "Error invalid signature")
	ErrUnknownPubKey        = tmsp.NewError(tmsp.CodeType_BaseUnknownPubKey, "Error unknown pubkey")

	ResultOK = tmsp.NewResultOK(nil, "")
)
