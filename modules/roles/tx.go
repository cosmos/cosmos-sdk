package roles

import (
	"github.com/tendermint/go-wire/data"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
)

const (
	// MaxMembers it the maximum number of members in a Role.  Used to avoid
	// extremely large roles.
	// Value is arbitrary, please adjust as needed
	MaxMembers = 20
)

// AssumeRoleTx is a layered tx that can wrap your normal tx to give it
// the authority to use a given role.
type AssumeRoleTx struct {
	Role data.Bytes  `json:"role"`
	Tx   basecoin.Tx `json:"tx"`
}

// NewAssumeRoleTx creates a new wrapper to add a role to a tx execution
func NewAssumeRoleTx(role []byte, tx basecoin.Tx) basecoin.Tx {
	return AssumeRoleTx{Role: role, Tx: tx}.Wrap()
}

// ValidateBasic - validate nothing is empty
func (tx AssumeRoleTx) ValidateBasic() error {
	if len(tx.Role) == 0 {
		return ErrNoRole()
	}
	if tx.Tx.Empty() {
		return errors.ErrUnknownTxType(tx.Tx)
	}
	return nil
}

// Wrap - used to satisfy TxInner
func (tx AssumeRoleTx) Wrap() basecoin.Tx {
	return basecoin.Tx{tx}
}

// CreateRoleTx is used to construct a new role
//
// TODO: add ability to update signers on a role... but that adds a lot
// more complexity to the permissions
type CreateRoleTx struct {
	Role    data.Bytes       `json:"role"`
	MinSigs uint32           `json:"min_sigs"`
	Signers []basecoin.Actor `json:"signers"`
}

// NewCreateRoleTx creates a new role, which we can later use
func NewCreateRoleTx(role []byte, minSigs uint32, signers []basecoin.Actor) basecoin.Tx {
	return CreateRoleTx{Role: role, MinSigs: minSigs, Signers: signers}.Wrap()
}

// ValidateBasic - validate nothing is empty
func (tx CreateRoleTx) ValidateBasic() error {
	if len(tx.Role) == 0 {
		return ErrNoRole()
	}
	if tx.MinSigs == 0 {
		return ErrNoMembers()
	}
	if len(tx.Signers) == 0 {
		return ErrNoMembers()
	}
	if len(tx.Signers) < int(tx.MinSigs) {
		return ErrNotEnoughMembers()
	}
	if len(tx.Signers) > MaxMembers {
		return ErrTooManyMembers()
	}
	return nil
}

// Wrap - used to satisfy TxInner
func (tx CreateRoleTx) Wrap() basecoin.Tx {
	return basecoin.Tx{tx}
}
