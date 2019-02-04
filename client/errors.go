package client

import "errors"

// common errors for CLI and REST clients
var (
	ErrInvalidGasAdjustment = errors.New("invalid gas adjustment")
	ErrInvalidSigner        = errors.New("tx intended signer does not match the given signer")
)
