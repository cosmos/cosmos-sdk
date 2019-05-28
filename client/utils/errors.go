package utils

import "errors"

var errInvalidSigner = errors.New("tx intended signer does not match the given signer")
