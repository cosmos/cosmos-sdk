package stack

import (
	"fmt"

	"github.com/tendermint/tmlibs/log"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/types"
)

const (
	NameDispatcher = "disp"
)

// Dispatcher grabs a bunch of Dispatchables and groups them into one Handler.
//
// It will route tx to the proper locations and also allows them to call each
// other synchronously through the same tx methods.
type Dispatcher struct {
	routes map[string]Dispatchable
}

func NewDispatcher(routes ...Dispatchable) *Dispatcher {
	d := &Dispatcher{}
	d.AddRoutes(routes...)
	return d
}

var _ basecoin.Handler = new(Dispatcher)

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

func (d *Dispatcher) Name() string {
	return NameDispatcher
}

func (d *Dispatcher) CheckTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
	r, err := d.lookupTx(tx)
	if err != nil {
		return res, err
	}
	// TODO: callback
	return r.CheckTx(ctx, store, tx, nil)
}

func (d *Dispatcher) DeliverTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
	r, err := d.lookupTx(tx)
	if err != nil {
		return res, err
	}
	// TODO: callback
	return r.DeliverTx(ctx, store, tx, nil)
}

func (d *Dispatcher) SetOption(l log.Logger, store types.KVStore, module, key, value string) (string, error) {
	r, err := d.lookupModule(module)
	if err != nil {
		return "", err
	}
	// TODO: callback
	return r.SetOption(l, store, module, key, value, nil)
}

func (d *Dispatcher) lookupTx(tx basecoin.Tx) (Dispatchable, error) {
	// TODO
	name := "foo"
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
