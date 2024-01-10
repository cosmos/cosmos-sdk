package mock

import (
	"cosmossdk.io/server/v2/core/transaction"
	"cosmossdk.io/server/v2/stf"
)

func STF[T transaction.Tx]() stf.STF[T] {
	panic("impl")
}
