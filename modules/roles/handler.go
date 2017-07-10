package roles

import (
	"github.com/tendermint/tmlibs/log"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/state"
)

//NameRole - name space of the roles module
const NameRole = "role"

type Handler struct{}

var _ basecoin.Handler = Handler{}

// NewHandler makes a role handler to create roles
func NewHandler() Handler {
	return Handler{}
}

// Name - return name space
func (Handler) Name() string {
	return NameRole
}

// CheckTx checks if there is enough money in the account
func (h Handler) CheckTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
	_, err = checkTx(ctx, tx)
	return res, err
}

// DeliverTx moves the money
func (h Handler) DeliverTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
	create, err := checkTx(ctx, tx)
	if err != nil {
		return res, err
	}

	// lets try...
	role := NewRole(create.MinSigs, create.Signers)
	err = createRole(store, MakeKey(create.Role), role)
	return res, err
}

// SetOption - sets the genesis account balance
func (h Handler) SetOption(l log.Logger, store state.KVStore, module, key, value string) (log string, err error) {
	if module != NameRole {
		return "", errors.ErrUnknownModule(module)
	}
	return "", errors.ErrUnknownKey(key)
}

func checkTx(ctx basecoin.Context, tx basecoin.Tx) (create CreateRoleTx, err error) {
	// check if the tx is proper type and valid
	create, ok := tx.Unwrap().(CreateRoleTx)
	if !ok {
		return create, errors.ErrInvalidFormat(tx)
	}
	err = create.ValidateBasic()
	return create, err
}
