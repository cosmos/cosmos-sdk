//nolint
package fee

import (
	"fmt"

	abci "github.com/tendermint/abci/types"

	"github.com/tendermint/basecoin/errors"
)

var (
	errInsufficientFees = fmt.Errorf("Insufficient Fees")
	errWrongFeeDenom    = fmt.Errorf("Required fee denomination")
)

func ErrInsufficientFees() errors.TMError {
	return errors.WithCode(errInsufficientFees, abci.CodeType_BaseInvalidInput)
}
func IsInsufficientFeesErr(err error) bool {
	return errors.IsSameError(errInsufficientFees, err)
}

func ErrWrongFeeDenom(denom string) errors.TMError {
	return errors.WithMessage(denom, errWrongFeeDenom, abci.CodeType_BaseInvalidInput)
}
func IsWrongFeeDenomErr(err error) bool {
	return errors.IsSameError(errWrongFeeDenom, err)
}
