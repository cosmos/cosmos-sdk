//nolint
package fee

import (
	"fmt"

	abci "github.com/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/errors"
)

var (
	errInsufficientFees = fmt.Errorf("Insufficient fees")
	errWrongFeeDenom    = fmt.Errorf("Required fee denomination")
	errSkipFees         = fmt.Errorf("Skip fees")

	invalidInput = abci.CodeType_BaseInvalidInput
)

func ErrInsufficientFees() errors.TMError {
	return errors.WithCode(errInsufficientFees, invalidInput)
}
func IsInsufficientFeesErr(err error) bool {
	return errors.IsSameError(errInsufficientFees, err)
}

func ErrWrongFeeDenom(denom string) errors.TMError {
	return errors.WithMessage(denom, errWrongFeeDenom, invalidInput)
}
func IsWrongFeeDenomErr(err error) bool {
	return errors.IsSameError(errWrongFeeDenom, err)
}

func ErrSkipFees() errors.TMError {
	return errors.WithCode(errSkipFees, invalidInput)
}
func IsSkipFeesErr(err error) bool {
	return errors.IsSameError(errSkipFees, err)
}
