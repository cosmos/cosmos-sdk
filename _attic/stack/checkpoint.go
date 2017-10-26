package stack

import (
	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/state"
)

//nolint
const (
	NameCheckpoint = "check"
)

// Checkpoint isolates all data store below this
type Checkpoint struct {
	OnCheck   bool
	OnDeliver bool
	PassInitState
	PassInitValidate
}

// Name of the module - fulfills Middleware interface
func (Checkpoint) Name() string {
	return NameCheckpoint
}

var _ Middleware = Checkpoint{}

// CheckTx reverts all data changes if there was an error
func (c Checkpoint) CheckTx(ctx sdk.Context, store state.SimpleDB, tx sdk.Tx, next sdk.Checker) (res sdk.CheckResult, err error) {
	if !c.OnCheck {
		return next.CheckTx(ctx, store, tx)
	}
	ps := store.Checkpoint()
	res, err = next.CheckTx(ctx, ps, tx)
	if err == nil {
		err = store.Commit(ps)
	}
	return res, err
}

// DeliverTx reverts all data changes if there was an error
func (c Checkpoint) DeliverTx(ctx sdk.Context, store state.SimpleDB, tx sdk.Tx, next sdk.Deliver) (res sdk.DeliverResult, err error) {
	if !c.OnDeliver {
		return next.DeliverTx(ctx, store, tx)
	}
	ps := store.Checkpoint()
	res, err = next.DeliverTx(ctx, ps, tx)
	if err == nil {
		err = store.Commit(ps)
	}
	return res, err
}
