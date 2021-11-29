package ormerrors

import "github.com/cosmos/cosmos-sdk/types/errors"

var codespace = "orm"

var (
	UnsupportedKeyField = errors.New(codespace, 1, "unsupported key field")
	BytesFieldTooLong   = errors.New(codespace, 2, "bytes field is longer than 255 bytes")
)
