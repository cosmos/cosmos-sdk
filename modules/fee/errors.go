//nolint
package fee

import (
	"fmt"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin/errors"
)

var (
	errInsufficientFees = fmt.Errorf("Insufficient Fees")
)

func ErrInsufficientFees() errors.TMError {
	return errors.WithCode(errInsufficientFees, abci.CodeType_BaseInvalidInput)
}
func IsInsufficientFeesErr(err error) bool {
	return errors.IsSameError(errInsufficientFees, err)
}
