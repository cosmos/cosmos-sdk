//nolint
package state

import (
	"fmt"

	abci "github.com/tendermint/abci/types"
	"github.com/cosmos/cosmos-sdk/errors"
)

var (
	errNotASubTransaction = fmt.Errorf("Not a sub-transaction")
)

func ErrNotASubTransaction() errors.TMError {
	return errors.WithCode(errNotASubTransaction, abci.CodeType_InternalError)
}
func IsNotASubTransactionErr(err error) bool {
	return errors.IsSameError(errNotASubTransaction, err)
}
