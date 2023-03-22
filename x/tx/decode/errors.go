package decode

import "cosmossdk.io/errors"

const (
	txCodespace = "tx"
)

var (
	ErrUnknownField = errors.Register(txCodespace, 2, "unknown protobuf field")
)
