package etc

import (
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/state"
	wire "github.com/tendermint/go-wire"
)

const (
	// Name of the module for registering it
	Name = "etc"

	// CostSet is the gas needed for the set operation
	CostSet uint64 = 10
	// CostRemove is the gas needed for the remove operation
	CostRemove = 10
)

// Handler allows us to set and remove data
type Handler struct {
	basecoin.NopInitState
	basecoin.NopInitValidate
}

var _ basecoin.Handler = Handler{}

// NewHandler makes a role handler to modify data
func NewHandler() Handler {
	return Handler{}
}

// Name - return name space
func (Handler) Name() string {
	return Name
}

// CheckTx verifies if the transaction is properly formated
func (h Handler) CheckTx(ctx basecoin.Context, store state.SimpleDB, tx basecoin.Tx) (res basecoin.CheckResult, err error) {
	err = tx.ValidateBasic()
	if err != nil {
		return
	}

	switch tx.Unwrap().(type) {
	case SetTx:
		res = basecoin.NewCheck(CostSet, "")
	case RemoveTx:
		res = basecoin.NewCheck(CostRemove, "")
	default:
		err = errors.ErrUnknownTxType(tx)
	}
	return
}

// DeliverTx tries to create a new role.
//
// Returns an error if the role already exists
func (h Handler) DeliverTx(ctx basecoin.Context, store state.SimpleDB, tx basecoin.Tx) (res basecoin.DeliverResult, err error) {
	err = tx.ValidateBasic()
	if err != nil {
		return
	}

	switch t := tx.Unwrap().(type) {
	case SetTx:
		res, err = h.doSetTx(ctx, store, t)
	case RemoveTx:
		res, err = h.doRemoveTx(ctx, store, t)
	default:
		err = errors.ErrUnknownTxType(tx)
	}
	return
}

// doSetTx write to the store, overwriting any previous value
func (h Handler) doSetTx(ctx basecoin.Context, store state.SimpleDB, tx SetTx) (res basecoin.DeliverResult, err error) {
	data := NewData(tx.Value, ctx.BlockHeight())
	store.Set(tx.Key, wire.BinaryBytes(data))
	return
}

// doRemoveTx deletes the value from the store and returns the last value
func (h Handler) doRemoveTx(ctx basecoin.Context, store state.SimpleDB, tx RemoveTx) (res basecoin.DeliverResult, err error) {
	// we set res.Data so it gets returned to the client over the abci interface
	res.Data = store.Get(tx.Key)
	if len(res.Data) != 0 {
		store.Remove(tx.Key)
	}
	return
}
