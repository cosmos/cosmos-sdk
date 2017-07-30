package base

import (
	"strings"

	abci "github.com/tendermint/abci/types"
	wire "github.com/tendermint/go-wire"
	"github.com/tendermint/go-wire/data"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/stack"
	"github.com/tendermint/basecoin/state"
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
func (Multiplexer) CheckTx(ctx basecoin.Context, store state.SimpleDB, tx basecoin.Tx, next basecoin.Checker) (res basecoin.CheckResult, err error) {
	if mtx, ok := tx.Unwrap().(*MultiTx); ok {
		return runAllChecks(ctx, store, mtx.Txs, next)
	}
	return next.CheckTx(ctx, store, tx)
}

// DeliverTx splits the input tx and checks them all - fulfills Middlware interface
func (Multiplexer) DeliverTx(ctx basecoin.Context, store state.SimpleDB, tx basecoin.Tx, next basecoin.Deliver) (res basecoin.DeliverResult, err error) {
	if mtx, ok := tx.Unwrap().(*MultiTx); ok {
		return runAllDelivers(ctx, store, mtx.Txs, next)
	}
	return next.DeliverTx(ctx, store, tx)
}

func runAllChecks(ctx basecoin.Context, store state.SimpleDB, txs []basecoin.Tx, next basecoin.Checker) (res basecoin.CheckResult, err error) {
	// store all results, unless anything errors
	rs := make([]basecoin.CheckResult, len(txs))
	for i, stx := range txs {
		rs[i], err = next.CheckTx(ctx, store, stx)
		if err != nil {
			return
		}
	}
	// now combine the results into one...
	return combineChecks(rs), nil
}

func runAllDelivers(ctx basecoin.Context, store state.SimpleDB, txs []basecoin.Tx, next basecoin.Deliver) (res basecoin.DeliverResult, err error) {
	// store all results, unless anything errors
	rs := make([]basecoin.DeliverResult, len(txs))
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
func combineChecks(all []basecoin.CheckResult) basecoin.CheckResult {
	datas := make([]data.Bytes, len(all))
	logs := make([]string, len(all))
	var allocated, payments uint
	for i, r := range all {
		datas[i] = r.Data
		logs[i] = r.Log
		allocated += r.GasAllocated
		payments += r.GasPayment
	}
	return basecoin.CheckResult{
		Data:         wire.BinaryBytes(datas),
		Log:          strings.Join(logs, "\n"),
		GasAllocated: allocated,
		GasPayment:   payments,
	}
}

// combines all data bytes as a go-wire array.
// joins all log messages with \n
func combineDelivers(all []basecoin.DeliverResult) basecoin.DeliverResult {
	datas := make([]data.Bytes, len(all))
	logs := make([]string, len(all))
	var used uint
	var diffs []*abci.Validator
	for i, r := range all {
		datas[i] = r.Data
		logs[i] = r.Log
		used += r.GasUsed
		if len(r.Diff) > 0 {
			diffs = append(diffs, r.Diff...)
		}
	}
	return basecoin.DeliverResult{
		Data:    wire.BinaryBytes(datas),
		Log:     strings.Join(logs, "\n"),
		GasUsed: used,
		Diff:    diffs,
	}
}
