package roles

import (
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/state"
)

//NameRole - name space of the roles module
const NameRole = "role"

// Handler allows us to create new roles
type Handler struct {
	basecoin.NopOption
}

var _ basecoin.Handler = Handler{}

// NewHandler makes a role handler to create roles
func NewHandler() Handler {
	return Handler{}
}

// Name - return name space
func (Handler) Name() string {
	return NameRole
}

// CheckTx verifies if the transaction is properly formated
func (h Handler) CheckTx(ctx basecoin.Context, store state.SimpleDB, tx basecoin.Tx) (res basecoin.Result, err error) {
	var cr CreateRoleTx
	cr, err = checkTx(ctx, tx)
	if err != nil {
		return
	}
	err = checkNoRole(store, cr.Role)
	return
}

// DeliverTx tries to create a new role.
//
// Returns an error if the role already exists
func (h Handler) DeliverTx(ctx basecoin.Context, store state.SimpleDB, tx basecoin.Tx) (res basecoin.Result, err error) {
	create, err := checkTx(ctx, tx)
	if err != nil {
		return res, err
	}

	// lets try...
	role := NewRole(create.MinSigs, create.Signers)
	err = createRole(store, create.Role, role)
	return res, err
}

func checkTx(ctx basecoin.Context, tx basecoin.Tx) (create CreateRoleTx, err error) {
	// check if the tx is proper type and valid
	create, ok := tx.Unwrap().(CreateRoleTx)
	if !ok {
		return create, errors.ErrInvalidFormat(TypeCreateRoleTx, tx)
	}
	err = create.ValidateBasic()
	return create, err
}
