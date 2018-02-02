package roles

import (
	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/errors"
	"github.com/cosmos/cosmos-sdk/state"
)

const (
	//NameRole - name space of the roles module
	NameRole = "role"
	// CostCreate is the cost to create a new role
	CostCreate = uint64(40)
	// CostAssume is the cost to assume a role as part of a tx
	CostAssume = uint64(5)
)

// Handler allows us to create new roles
type Handler struct {
	sdk.NopInitState
	sdk.NopInitValidate
}

var _ sdk.Handler = Handler{}

// NewHandler makes a role handler to create roles
func NewHandler() Handler {
	return Handler{}
}

// Name - return name space
func (Handler) Name() string {
	return NameRole
}

// CheckTx verifies if the transaction is properly formated
func (h Handler) CheckTx(ctx sdk.Context, store state.SimpleDB, tx sdk.Tx) (res sdk.CheckResult, err error) {
	var cr CreateRoleTx
	cr, err = checkTx(ctx, tx)
	if err != nil {
		return
	}
	res = sdk.NewCheck(CostCreate, "")
	err = checkNoRole(store, cr.Role)
	return
}

// DeliverTx tries to create a new role.
//
// Returns an error if the role already exists
func (h Handler) DeliverTx(ctx sdk.Context, store state.SimpleDB, tx sdk.Tx) (res sdk.DeliverResult, err error) {
	create, err := checkTx(ctx, tx)
	if err != nil {
		return res, err
	}

	// lets try...
	role := NewRole(create.MinSigs, create.Signers)
	err = createRole(store, create.Role, role)
	return res, err
}

func checkTx(ctx sdk.Context, tx sdk.Tx) (create CreateRoleTx, err error) {
	// check if the tx is proper type and valid
	create, ok := tx.Unwrap().(CreateRoleTx)
	if !ok {
		return create, errors.ErrInvalidFormat(TypeCreateRoleTx, tx)
	}
	err = create.ValidateBasic()
	return create, err
}
