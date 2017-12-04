// XXX Rename AppHandler to DefaultAppHandler.
// XXX Register with a sdk.BaseApp instance to create Basecoin.
// XXX Create TxParser in anotehr file.

package app

import (
	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/store"
)

// AppHandler has no state for now, a more complex app could store state here
type AppHandler struct{}

func NewAppHandler() sdk.Handler {
	return AppHandler{}
}

// DeliverTx applies the tx
func (h Handler) DeliverTx(ctx sdk.Context, store store.MultiStore,
	msg interface{}) (res sdk.DeliverResult, err error) {

	db := store.Get("main").(sdk.KVStore)

	// Here we switch on which implementation of tx we use,
	// and then take the appropriate action.
	switch tx := tx.(type) {
	case SendTx:
		err = tx.ValidateBasic()
		if err != nil {
			break
		}
		db.Set(tx.Key, tx.Value)
		res.Data = tx.Key
	case RemoveTx:
		err = tx.ValidateBasic()
		if err != nil {
			break
		}
		db.Remove(tx.Key)
		res.Data = tx.Key
	default:
		err = errors.ErrInvalidFormat(TxWrapper{}, msg)
	}

	return
}

// CheckTx verifies if it is legit and returns info on how
// to prioritize it in the mempool
func (h Handler) CheckTx(ctx sdk.Context, store store.MultiStore,
	msg interface{}) (res sdk.CheckResult, err error) {

	// If we wanted to use the store,
	// it would look the same (as DeliverTx)
	// db := store.Get("main").(sdk.KVStore)

	// Make sure it is something valid
	tx, ok := msg.(Tx)
	if !ok {
		return res, errors.ErrInvalidFormat(TxWrapper{}, msg)
	}
	err = tx.ValidateBasic()
	if err != nil {
		return
	}

	// Now return the costs (these should have meaning in your app)
	return sdk.CheckResult{
		GasAllocated: 50,
		GasPayment:   10,
	}
}
