package base

import (
	"strings"

	abci "github.com/tendermint/abci/types"
	wire "github.com/tendermint/go-wire"
	"github.com/tendermint/go-wire/data"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/stack"
	"github.com/cosmos/cosmos-sdk/state"
)

//nolint
const (
	NameMultiplexer = "mplx"
)

// Multiplexer grabs a MultiTx and sends them sequentially down the line
type Multiplexer struct {
	stack.PassInitState
	stack.PassInitValidate
}

// Name of the module - fulfills Middleware interface
func (Multiplexer) Name() string {
	return NameMultiplexer
}

var _ stack.Middleware = Multiplexer{}

// CheckTx splits the input tx and checks them all - fulfills Middlware interface
func (Multiplexer) CheckTx(ctx sdk.Context, store state.SimpleDB, tx sdk.Tx, next sdk.Checker) (res sdk.CheckResult, err error) {
	if mtx, ok := tx.Unwrap().(MultiTx); ok {
		return runAllChecks(ctx, store, mtx.Txs, next)
	}
	return next.CheckTx(ctx, store, tx)
}

// DeliverTx splits the input tx and checks them all - fulfills Middlware interface
func (Multiplexer) DeliverTx(ctx sdk.Context, store state.SimpleDB, tx sdk.Tx, next sdk.Deliver) (res sdk.DeliverResult, err error) {
	if mtx, ok := tx.Unwrap().(MultiTx); ok {
		return runAllDelivers(ctx, store, mtx.Txs, next)
	}
	return next.DeliverTx(ctx, store, tx)
}

func runAllChecks(ctx sdk.Context, store state.SimpleDB, txs []sdk.Tx, next sdk.Checker) (res sdk.CheckResult, err error) {
	// store all results, unless anything errors
	rs := make([]sdk.CheckResult, len(txs))
	for i, stx := range txs {
		rs[i], err = next.CheckTx(ctx, store, stx)
		if err != nil {
			return
		}
	}
	// now combine the results into one...
	return combineChecks(rs), nil
}

func runAllDelivers(ctx sdk.Context, store state.SimpleDB, txs []sdk.Tx, next sdk.Deliver) (res sdk.DeliverResult, err error) {
	// store all results, unless anything errors
	rs := make([]sdk.DeliverResult, len(txs))
	for i, stx := range txs {
		rs[i], err = next.DeliverTx(ctx, store, stx)
		if err != nil {
			return
		}
	}
	// now combine the results into one...
	return combineDelivers(rs), nil
}

// combines all data bytes as a go-wire array.
// joins all log messages with \n
func combineChecks(all []sdk.CheckResult) sdk.CheckResult {
	datas := make([]data.Bytes, len(all))
	logs := make([]string, len(all))
	var allocated, payments uint64
	for i, r := range all {
		datas[i] = r.Data
		logs[i] = r.Log
		allocated += r.GasAllocated
		payments += r.GasPayment
	}
	return sdk.CheckResult{
		Data:         wire.BinaryBytes(datas),
		Log:          strings.Join(logs, "\n"),
		GasAllocated: allocated,
		GasPayment:   payments,
	}
}

// combines all data bytes as a go-wire array.
// joins all log messages with \n
func combineDelivers(all []sdk.DeliverResult) sdk.DeliverResult {
	datas := make([]data.Bytes, len(all))
	logs := make([]string, len(all))
	var used uint64
	var diffs []*abci.Validator
	for i, r := range all {
		datas[i] = r.Data
		logs[i] = r.Log
		used += r.GasUsed
		if len(r.Diff) > 0 {
			diffs = append(diffs, r.Diff...)
		}
	}
	return sdk.DeliverResult{
		Data:    wire.BinaryBytes(datas),
		Log:     strings.Join(logs, "\n"),
		GasUsed: used,
		Diff:    diffs,
	}
}
