package types

import "errors"

var (
	ErrorInvalidSigner        = errors.New("tx intended signer does not match the given signer")
	ErrorInvalidGasAdjustment = errors.New("invalid gas adjustment")
)
