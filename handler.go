package sdk

import (
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-wire/data"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/state"
)

const (
	// ModuleNameBase is the module name for internal functionality
	ModuleNameBase = "base"
	// ChainKey is the option key for setting the chain id
	ChainKey = "chain_id"
)

// Handler is anything that processes a transaction
type Handler interface {
	// Checker verifies there are valid fees and estimates work
	Checker
	// Deliver performs the tx once it makes it in the block
	Deliver
	// InitStater sets state from the genesis file
	InitStater
	// InitValidater sets the initial validator set
	InitValidater
	// Named ensures there is a name for the item
	Named

	// TODO????
	// BeginBlock(store state.SimpleDB, hash []byte, header *abci.Header)
}

// Ticker can be executed every block
type Ticker interface {
	Tick(Context, state.SimpleDB) ([]*abci.Validator, error)
}

// TickerFunc allows a function to implement the interface
type TickerFunc func(Context, state.SimpleDB) ([]*abci.Validator, error)

func (t TickerFunc) Tick(ctx Context, store state.SimpleDB) ([]*abci.Validator, error) {
	return t(ctx, store)
}

// Named ensures there is a name for the item
type Named interface {
	Name() string
}

// Checker verifies there are valid fees and estimates work
type Checker interface {
	CheckTx(ctx Context, store state.SimpleDB, tx Tx) (CheckResult, error)
}

// CheckerFunc (like http.HandlerFunc) is a shortcut for making wrappers
type CheckerFunc func(Context, state.SimpleDB, Tx) (CheckResult, error)

func (c CheckerFunc) CheckTx(ctx Context, store state.SimpleDB, tx Tx) (CheckResult, error) {
	return c(ctx, store, tx)
}

// Deliver performs the tx once it makes it in the block
type Deliver interface {
	DeliverTx(ctx Context, store state.SimpleDB, tx Tx) (DeliverResult, error)
}

// DeliverFunc (like http.HandlerFunc) is a shortcut for making wrappers
type DeliverFunc func(Context, state.SimpleDB, Tx) (DeliverResult, error)

func (c DeliverFunc) DeliverTx(ctx Context, store state.SimpleDB, tx Tx) (DeliverResult, error) {
	return c(ctx, store, tx)
}

// InitStater sets state from the genesis file
type InitStater interface {
	InitState(l log.Logger, store state.SimpleDB, module, key, value string) (string, error)
}

// InitStateFunc (like http.HandlerFunc) is a shortcut for making wrappers
type InitStateFunc func(log.Logger, state.SimpleDB, string, string, string) (string, error)

func (c InitStateFunc) InitState(l log.Logger, store state.SimpleDB, module, key, value string) (string, error) {
	return c(l, store, module, key, value)
}

// InitValidater sets the initial validator set
type InitValidater interface {
	InitValidate(log log.Logger, store state.SimpleDB, vals []*abci.Validator)
}

// InitValidateFunc (like http.HandlerFunc) is a shortcut for making wrappers
type InitValidateFunc func(log.Logger, state.SimpleDB, []*abci.Validator)

func (c InitValidateFunc) InitValidate(l log.Logger, store state.SimpleDB, vals []*abci.Validator) {
	c(l, store, vals)
}

//---------- results and some wrappers --------

type Result interface {
	GetData() data.Bytes
}

// CheckResult captures any non-error abci result
// to make sure people use error for error cases
type CheckResult struct {
	Data data.Bytes
	Log  string
	// GasAllocated is the maximum units of work we allow this tx to perform
	GasAllocated int64
	// GasPayment is the total fees for this tx (or other source of payment)
	GasPayment int64
}

// NewCheck sets the gas used and the response data but no more info
// these are the most common info needed to be set by the Handler
func NewCheck(gasAllocated int64, log string) CheckResult {
	return CheckResult{
		GasAllocated: gasAllocated,
		Log:          log,
	}
}

func (c CheckResult) ToABCI() abci.ResponseCheckTx {
	return abci.ResponseCheckTx{
		Data:      c.Data,
		Log:       c.Log,
		GasWanted: c.GasAllocated,
		Fee:       cmn.KI64Pair{[]byte("gas"), c.GasPayment},
	}
}

func (c CheckResult) GetData() data.Bytes {
	return c.Data
}

// DeliverResult captures any non-error abci result
// to make sure people use error for error cases
type DeliverResult struct {
	Data    data.Bytes
	Log     string
	Diff    []*abci.Validator
	Tags    []cmn.KVPair
	GasUsed int64 // unused
}

func (d DeliverResult) ToABCI() abci.ResponseDeliverTx {
	return abci.ResponseDeliverTx{
		Data: d.Data,
		Log:  d.Log,
		Tags: d.Tags,
	}
}

func (d DeliverResult) GetData() data.Bytes {
	return d.Data
}

// placeholders
// holders
type NopCheck struct{}

func (_ NopCheck) CheckTx(Context, state.SimpleDB, Tx) (r CheckResult, e error) { return }

type NopDeliver struct{}

func (_ NopDeliver) DeliverTx(Context, state.SimpleDB, Tx) (r DeliverResult, e error) { return }

type NopInitState struct{}

func (_ NopInitState) InitState(log.Logger, state.SimpleDB, string, string, string) (string, error) {
	return "", nil
}

type NopInitValidate struct{}

func (_ NopInitValidate) InitValidate(log log.Logger, store state.SimpleDB, vals []*abci.Validator) {}
