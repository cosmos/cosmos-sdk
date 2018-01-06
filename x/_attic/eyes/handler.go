package eyes

import (
	wire "github.com/tendermint/go-wire"
	"github.com/tendermint/tmlibs/log"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/errors"
)

const (
	// Name is used to register this module
	Name = "eyes"
	// CostSet is the gas needed for the set operation
	CostSet uint64 = 10
	// CostRemove is the gas needed for the remove operation
	CostRemove = 10
)

// Handler allows us to set and remove data
type Handler struct{}

var _ sdk.Handler = Handler{}

// NewHandler makes a role handler to modify data
func NewHandler() Handler {
	return Handler{}
}

// InitState - sets the genesis state - implements InitStater
func (h Handler) InitState(l log.Logger, store sdk.SimpleDB,
	module, key, value string) (log string, err error) {

	if module != Name {
		return "", errors.ErrUnknownModule(module)
	}
	store.Set([]byte(key), []byte(value))
	return key, nil
}

// CheckTx verifies if the transaction is properly formated
func (h Handler) CheckTx(ctx sdk.Context, store sdk.SimpleDB,
	msg interface{}) (res sdk.CheckResult, err error) {

	tx := sdk.MustGetTx(msg).(EyesTx)
	if err := tx.ValidateBasic(); err != nil {
		return res, err
	}

	switch tx.Unwrap().(type) {
	case SetTx:
		res = sdk.NewCheck(CostSet, "")
	case RemoveTx:
		res = sdk.NewCheck(CostRemove, "")
	default:
		err = errors.ErrUnknownTxType(tx)
	}
	return
}

// DeliverTx tries to create a new role.
//
// Returns an error if the role already exists
func (h Handler) DeliverTx(ctx sdk.Context, store sdk.SimpleDB,
	msg interface{}) (res sdk.DeliverResult, err error) {

	tx := sdk.MustGetTx(msg).(EyesTx)
	if err := tx.ValidateBasic(); err != nil {
		return res, err
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

// doSetTx writes to the store, overwriting any previous value
// note that an empty response in DeliverTx is OK with no log or data returned
func (h Handler) doSetTx(ctx sdk.Context, store sdk.SimpleDB,
	tx SetTx) (res sdk.DeliverResult, err error) {

	data := NewData(tx.Value, ctx.BlockHeight())
	store.Set(tx.Key, wire.BinaryBytes(data))
	return
}

// doRemoveTx deletes the value from the store and returns the last value
// here we let res.Data to return the value over abci
func (h Handler) doRemoveTx(ctx sdk.Context, store sdk.SimpleDB,
	tx RemoveTx) (res sdk.DeliverResult, err error) {

	// we set res.Data so it gets returned to the client over the abci interface
	res.Data = store.Get(tx.Key)
	if len(res.Data) != 0 {
		store.Remove(tx.Key)
	}
	return
}
