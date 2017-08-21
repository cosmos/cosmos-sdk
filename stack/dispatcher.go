package stack

import (
	"fmt"
	"sort"
	"strings"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/tmlibs/log"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/errors"
	"github.com/cosmos/cosmos-sdk/state"
)

// nolint
const (
	NameDispatcher = "disp"
)

// Dispatcher grabs a bunch of Dispatchables and groups them into one Handler.
//
// It will route tx to the proper locations and also allows them to call each
// other synchronously through the same tx methods.
//
// Please note that iterating through a map is a non-deteministic operation
// and, as such, should never be done in the context of an ABCI app.  Only
// use this map to look up an exact route by name.
type Dispatcher struct {
	routes map[string]Dispatchable
}

// NewDispatcher creates a dispatcher and adds the given routes.
// You can also add routes later with .AddRoutes()
func NewDispatcher(routes ...Dispatchable) *Dispatcher {
	d := &Dispatcher{
		routes: map[string]Dispatchable{},
	}
	d.AddRoutes(routes...)
	return d
}

var _ sdk.Handler = new(Dispatcher)

// AddRoutes registers all these dispatchable choices under their subdomains
//
// Panics on attempt to double-register a route name, as this is a configuration error.
// Should I retrun an error instead?
func (d *Dispatcher) AddRoutes(routes ...Dispatchable) {
	for _, r := range routes {
		name := r.Name()
		if _, ok := d.routes[name]; ok {
			panic(fmt.Sprintf("%s already registered with dispatcher", name))
		}
		d.routes[name] = r
	}
}

// Name - defines the name of this module
func (d *Dispatcher) Name() string {
	return NameDispatcher
}

// CheckTx - implements Handler interface
//
// Tries to find a registered module (Dispatchable) based on the name of the tx.
// The tx name (as registered with go-data) should be in the form `<module name>/XXXX`,
// where `module name` must match the name of a dispatchable and XXX can be any string.
func (d *Dispatcher) CheckTx(ctx sdk.Context, store state.SimpleDB, tx sdk.Tx) (res sdk.CheckResult, err error) {
	r, err := d.lookupTx(tx)
	if err != nil {
		return res, err
	}

	// make sure no monkey business with the context
	cb := secureCheck(d, ctx)

	// and isolate the permissions and the data store for this app
	ctx = withApp(ctx, r.Name())
	store = stateSpace(store, r.Name())

	return r.CheckTx(ctx, store, tx, cb)
}

// DeliverTx - implements Handler interface
//
// Tries to find a registered module (Dispatchable) based on the name of the tx.
// The tx name (as registered with go-data) should be in the form `<module name>/XXXX`,
// where `module name` must match the name of a dispatchable and XXX can be any string.
func (d *Dispatcher) DeliverTx(ctx sdk.Context, store state.SimpleDB, tx sdk.Tx) (res sdk.DeliverResult, err error) {
	r, err := d.lookupTx(tx)
	if err != nil {
		return res, err
	}

	// make sure no monkey business with the context
	cb := secureDeliver(d, ctx)

	// and isolate the permissions and the data store for this app
	ctx = withApp(ctx, r.Name())
	store = stateSpace(store, r.Name())

	return r.DeliverTx(ctx, store, tx, cb)
}

// InitState - implements Handler interface
//
// Tries to find a registered module (Dispatchable) based on the
// module name from InitState of the tx.
func (d *Dispatcher) InitState(l log.Logger, store state.SimpleDB, module, key, value string) (string, error) {
	r, err := d.lookupModule(module)
	if err != nil {
		return "", err
	}

	// no ctx, so secureCheck not needed
	cb := d
	// but isolate data space
	store = stateSpace(store, r.Name())

	return r.InitState(l, store, module, key, value, cb)
}

// InitValidate makes sure all modules are informed
func (d *Dispatcher) InitValidate(log log.Logger, store state.SimpleDB, vals []*abci.Validator) {
	for _, mod := range d.sortedModules() {
		// no ctx, so secureCheck not needed
		cb := d
		space := stateSpace(store, mod.Name())
		mod.InitValidate(log, space, vals, cb)
	}
}

func (d *Dispatcher) lookupTx(tx sdk.Tx) (Dispatchable, error) {
	kind, err := tx.GetKind()
	if err != nil {
		return nil, err
	}
	// grab everything before the /
	name := strings.SplitN(kind, "/", 2)[0]
	r, ok := d.routes[name]
	if !ok {
		return nil, errors.ErrUnknownTxType(tx)
	}
	return r, nil
}

func (d *Dispatcher) lookupModule(name string) (Dispatchable, error) {
	r, ok := d.routes[name]
	if !ok {
		return nil, errors.ErrUnknownModule(name)
	}
	return r, nil
}

func (d *Dispatcher) sortedModules() []Dispatchable {
	// order all routes names
	size := len(d.routes)
	names := make([]string, 0, size)
	for k := range d.routes {
		names = append(names, k)
	}
	sort.Strings(names)

	res := make([]Dispatchable, size)
	for i, k := range names {
		res[i] = d.routes[k]
	}
	return res
}
