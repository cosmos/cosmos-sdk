package types

import (
	sdkerrors "cosmossdk.io/errors"
	abci "github.com/tendermint/tendermint/abci/types"
)

const StoreCodespace = "store"

var (
	// ErrInvalidProof is returned when a proof is invalid
	ErrInvalidProof = sdkerrors.Register(StoreCodespace, 2, "invalid proof")
	// ErrTxDecode is returned if we cannot parse a transaction
	ErrTxDecode = sdkerrors.Register(StoreCodespace, 3, "tx parse error")

	// ErrUnknownRequest to doc
	ErrUnknownRequest = sdkerrors.Register(StoreCodespace, 4, "unknown request")

	// ErrLogic defines an internal logic error, e.g. an invariant or assertion
	// that is violated. It is a programmer error, not a user-facing error.
	ErrLogic = sdkerrors.Register(StoreCodespace, 5, "internal logic error")

	// ErrConflict defines a conflict error, e.g. when two goroutines try to access
	// the same resource and one of them fails.
	ErrConflict = sdkerrors.Register(StoreCodespace, 6, "conflict")
	// ErrInvalidRequest defines an ABCI typed error where the request contains
	// invalid data.
	ErrInvalidRequest = sdkerrors.Register(StoreCodespace, 7, "invalid request")
)

// ABCI QueryResult

// QueryResult returns a ResponseQuery from an error. It will try to parse ABCI
// info from the error.
func QueryResult(err error, debug bool) abci.ResponseQuery {
	space, code, log := sdkerrors.ABCIInfo(err, debug)
	return abci.ResponseQuery{
		Codespace: space,
		Code:      code,
		Log:       log,
	}
}
