package sdk

import (
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/state"
)

const (
	// ModuleNameBase is the module name for internal functionality
	ModuleNameBase = "base"
	// ChainKey is the option key for setting the chain id
	ChainKey = "chain_id"
)

// Handler is anything that processes a transaction.
// Must handle checktx and delivertx
type Handler interface {
	// Checker verifies there are valid fees and estimates work
	Checker
	// Deliver performs the tx once it makes it in the block
	Deliverer
}

// Checker verifies there are valid fees and estimates work
type Checker interface {
	CheckTx(ctx Context, store state.SimpleDB, tx interface{}) (CheckResult, error)
}

// CheckerFunc (like http.HandlerFunc) is a shortcut for making wrappers
type CheckerFunc func(Context, state.SimpleDB, interface{}) (CheckResult, error)

func (c CheckerFunc) CheckTx(ctx Context, store state.SimpleDB, tx interface{}) (CheckResult, error) {
	return c(ctx, store, tx)
}

// Deliverer performs the tx once it makes it in the block
type Deliverer interface {
	DeliverTx(ctx Context, store state.SimpleDB, tx interface{}) (DeliverResult, error)
}

// DelivererFunc (like http.HandlerFunc) is a shortcut for making wrappers
type DelivererFunc func(Context, state.SimpleDB, interface{}) (DeliverResult, error)

func (c DelivererFunc) DeliverTx(ctx Context, store state.SimpleDB, tx interface{}) (DeliverResult, error) {
	return c(ctx, store, tx)
}

/////////////////////////////////////////////////
// Lifecycle actions, not tied to the tx handler

// Ticker can be executed every block.
// Called from BeginBlock
type Ticker interface {
	Tick(Context, state.SimpleDB) ([]*abci.Validator, error)
}

// TickerFunc allows a function to implement the interface
type TickerFunc func(Context, state.SimpleDB) ([]*abci.Validator, error)

func (t TickerFunc) Tick(ctx Context, store state.SimpleDB) ([]*abci.Validator, error) {
	return t(ctx, store)
}

// InitValidator sets the initial validator set.
// Called from InitChain
type InitValidator interface {
	InitValidators(logger log.Logger, store state.SimpleDB,
		vals []*abci.Validator)
}

// InitStater sets state from the genesis file
//
// TODO: Think if this belongs here, in genesis, or somewhere else
type InitStater interface {
	InitState(logger log.Logger, store state.SimpleDB,
		module, key, value string) (string, error)
}

//////////////////////////////////////////////////
// Helper methods

// Msg allows us to get the actual tx from a structure with lots of
// decorator information. This is usually what should be passed to Handlers.
type Msg interface {
	GetTx() interface{}
}

// MustGetTx forces the msg to the interface and extracts the tx
func MustGetTx(msg interface{}) interface{} {
	m := msg.(Msg)
	return m.GetTx()
}
