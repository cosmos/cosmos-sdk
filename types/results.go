package types

import (
	abci "github.com/tendermint/abci/types"
)

// CheckResult captures any non-error ABCI  result
// to make sure people use error for error cases.
type CheckResult struct {
	abci.Result

	// GasAllocated is the maximum units of work we allow this tx to perform
	GasAllocated uint64

	// GasPayment is the total fees for this tx (or other source of payment)
	GasPayment uint64
}

// DeliverResult captures any non-error abci result
// to make sure people use error for error cases
type DeliverResult struct {
	abci.Result

	// TODO comment
	Diff []*abci.Validator

	// TODO comment
	GasUsed uint64
}
