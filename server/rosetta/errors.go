package rosetta

import "github.com/tendermint/cosmos-rosetta-gateway/rosetta"

var (
	ErrInterpreting                = rosetta.NewError(1, "error interpreting data from node")
	ErrNodeConnection              = rosetta.NewError(2, "error getting data from node")
	ErrUnsupportedCurve            = rosetta.NewError(3, "unsupported curve, expected secp256k1")
	ErrInvalidOperation            = rosetta.NewError(4, "invalid operation")
	ErrInvalidTransaction          = rosetta.NewError(5, "invalid transaction")
	ErrInvalidRequest              = rosetta.NewError(6, "invalid request")
	ErrInvalidAddress              = rosetta.NewError(7, "invalid address")
	ErrInvalidPubkey               = rosetta.NewError(8, "invalid pubkey")
	ErrEndpointDisabledOfflineMode = rosetta.NewError(9, "endpoint disabled in offline mode")
	ErrInvalidTxHash               = rosetta.NewError(10, "invalid tx hash")
)
