package basecoin

import (
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-wire/data"
	"github.com/tendermint/tmlibs/log"

	"github.com/tendermint/basecoin/state"
)

// Handler is anything that processes a transaction
type Handler interface {
	Checker
	Deliver
	SetOptioner
	Named
	// TODO: flesh these out as well
	// InitChain(store state.SimpleDB, vals []*abci.Validator)
	// BeginBlock(store state.SimpleDB, hash []byte, header *abci.Header)
	// EndBlock(store state.SimpleDB, height uint64) abci.ResponseEndBlock
}

type Named interface {
	Name() string
}

type Checker interface {
	CheckTx(ctx Context, store state.SimpleDB, tx Tx) (CheckResult, error)
}

// CheckerFunc (like http.HandlerFunc) is a shortcut for making wrapers
type CheckerFunc func(Context, state.SimpleDB, Tx) (CheckResult, error)

func (c CheckerFunc) CheckTx(ctx Context, store state.SimpleDB, tx Tx) (CheckResult, error) {
	return c(ctx, store, tx)
}

type Deliver interface {
	DeliverTx(ctx Context, store state.SimpleDB, tx Tx) (DeliverResult, error)
}

// DeliverFunc (like http.HandlerFunc) is a shortcut for making wrapers
type DeliverFunc func(Context, state.SimpleDB, Tx) (DeliverResult, error)

func (c DeliverFunc) DeliverTx(ctx Context, store state.SimpleDB, tx Tx) (DeliverResult, error) {
	return c(ctx, store, tx)
}

type SetOptioner interface {
	SetOption(l log.Logger, store state.SimpleDB, module, key, value string) (string, error)
}

// SetOptionFunc (like http.HandlerFunc) is a shortcut for making wrapers
type SetOptionFunc func(log.Logger, state.SimpleDB, string, string, string) (string, error)

func (c SetOptionFunc) SetOption(l log.Logger, store state.SimpleDB, module, key, value string) (string, error) {
	return c(l, store, module, key, value)
}

//---------- results and some wrappers --------

type Dataer interface {
	GetData() data.Bytes
}

// CheckResult captures any non-error abci result
// to make sure people use error for error cases
type CheckResult struct {
	Data         data.Bytes
	Log          string
	GasAllocated uint
	GasPrice     uint
}

var _ Dataer = CheckResult{}

func (r CheckResult) ToABCI() abci.Result {
	return abci.Result{
		Data: r.Data,
		Log:  r.Log,
	}
}

func (r CheckResult) GetData() data.Bytes {
	return r.Data
}

// DeliverResult captures any non-error abci result
// to make sure people use error for error cases
type DeliverResult struct {
	Data    data.Bytes
	Log     string
	Diff    []*abci.Validator
	GasUsed uint
}

var _ Dataer = DeliverResult{}

func (r DeliverResult) ToABCI() abci.Result {
	return abci.Result{
		Data: r.Data,
		Log:  r.Log,
	}
}

func (r DeliverResult) GetData() data.Bytes {
	return r.Data
}

// placeholders
// holders
type NopCheck struct{}

func (_ NopCheck) CheckTx(Context, state.SimpleDB, Tx) (r CheckResult, e error) { return }

type NopDeliver struct{}

func (_ NopDeliver) DeliverTx(Context, state.SimpleDB, Tx) (r DeliverResult, e error) { return }

type NopOption struct{}

func (_ NopOption) SetOption(log.Logger, state.SimpleDB, string, string, string) (string, error) {
	return "", nil
}
