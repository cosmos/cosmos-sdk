//nolint
package nonce

import (
	"fmt"

	abci "github.com/tendermint/abci/types"

	"github.com/tendermint/basecoin/errors"
)

var (
	errNoNonce      = fmt.Errorf("Tx doesn't contain nonce")
	errNotMember    = fmt.Errorf("nonce contains non-permissioned member")
	errZeroSequence = fmt.Errorf("Sequence number cannot be zero")

	unauthorized = abci.CodeType_Unauthorized
)

func ErrBadNonce(got, expected uint32) errors.TMError {
	return errors.WithCode(fmt.Errorf("Bad nonce sequence, got %d, expected %d", got, expected), unauthorized)
}

func ErrNoNonce() errors.TMError {
	return errors.WithCode(errNoNonce, unauthorized)
}
func ErrNotMember() errors.TMError {
	return errors.WithCode(errNotMember, unauthorized)
}
func ErrZeroSequence() errors.TMError {
	return errors.WithCode(errZeroSequence, unauthorized)
}
