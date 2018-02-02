package roles

import (
	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/stack"
	"github.com/cosmos/cosmos-sdk/state"
)

// Middleware allows us to add a requested role as a permission
// if the tx requests it and has sufficient authority
type Middleware struct {
	stack.PassInitState
	stack.PassInitValidate
}

var _ stack.Middleware = Middleware{}

// NewMiddleware creates a role-checking middleware
func NewMiddleware() Middleware {
	return Middleware{}
}

// Name - return name space
func (Middleware) Name() string {
	return NameRole
}

// CheckTx tries to assume the named role if requested.
// If no role is requested, do nothing.
// If insufficient authority to assume the role, return error.
func (m Middleware) CheckTx(ctx sdk.Context, store state.SimpleDB, tx sdk.Tx, next sdk.Checker) (res sdk.CheckResult, err error) {
	// if this is not an AssumeRoleTx, then continue
	assume, ok := tx.Unwrap().(AssumeRoleTx)
	if !ok { // this also breaks the recursion below
		return next.CheckTx(ctx, store, tx)
	}

	ctx, err = assumeRole(ctx, store, assume)
	if err != nil {
		return res, err
	}

	// one could add multiple role statements, repeat as needed
	// charging for each level
	res, err = m.CheckTx(ctx, store, assume.Tx, next)
	res.GasAllocated += CostAssume
	return
}

// DeliverTx tries to assume the named role if requested.
// If no role is requested, do nothing.
// If insufficient authority to assume the role, return error.
func (m Middleware) DeliverTx(ctx sdk.Context, store state.SimpleDB, tx sdk.Tx, next sdk.Deliver) (res sdk.DeliverResult, err error) {
	// if this is not an AssumeRoleTx, then continue
	assume, ok := tx.Unwrap().(AssumeRoleTx)
	if !ok { // this also breaks the recursion below
		return next.DeliverTx(ctx, store, tx)
	}

	ctx, err = assumeRole(ctx, store, assume)
	if err != nil {
		return res, err
	}

	// one could add multiple role statements, repeat as needed
	return m.DeliverTx(ctx, store, assume.Tx, next)
}

func assumeRole(ctx sdk.Context, store state.SimpleDB, assume AssumeRoleTx) (sdk.Context, error) {
	err := assume.ValidateBasic()
	if err != nil {
		return nil, err
	}

	role, err := loadRole(store, assume.Role)
	if err != nil {
		return nil, err
	}

	if !role.IsAuthorized(ctx) {
		return nil, ErrInsufficientSigs()
	}
	ctx = ctx.WithPermissions(NewPerm(assume.Role))
	return ctx, nil
}
