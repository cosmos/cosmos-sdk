package auth

import (
	sdk "github.com/cosmos/cosmos-sdk"
)

type CheckSignatures struct{}

var _ sdk.Decorator = CheckSignatures{}

// CheckTx verifies the signatures are correct - fulfills Middlware interface
func (Signatures) CheckTx(ctx sdk.Context, store sdk.SimpleDB,
	tx interface{}, next sdk.Checker) (res sdk.CheckResult, err error) {

	// Check that signatures match
	// TODO

	// Add info to context
	// TODO

	return next.CheckTx(ctx2, store, tx)
}

// DeliverTx verifies the signatures are correct - fulfills Middlware interface
func (Signatures) DeliverTx(ctx sdk.Context, store sdk.SimpleDB,
	tx interface{}, next sdk.Deliverer) (res sdk.DeliverResult, err error) {

	// Check that signatures match
	// TODO

	// Add info to context
	// TODO

	return next.DeliverTx(ctx2, store, tx)
}
