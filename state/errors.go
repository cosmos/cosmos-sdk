//nolint
package state

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/errors"
)

var (
	errNotASubTransaction = fmt.Errorf("Not a sub-transaction")
)

func ErrNotASubTransaction() errors.TMError {
	return errors.WithCode(errNotASubTransaction, errors.CodeTypeInternalErr)
}
func IsNotASubTransactionErr(err error) bool {
	return errors.IsSameError(errNotASubTransaction, err)
}
