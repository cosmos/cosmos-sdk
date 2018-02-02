package roles

import (
	"github.com/tendermint/go-wire/data"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/errors"
)

var (
	// MaxMembers it the maximum number of members in a Role.  Used to avoid
	// extremely large roles.
	// Value is arbitrary, please adjust as needed
	MaxMembers = 20
)

//nolint
const (
	ByteAssumeRoleTx = 0x23
	ByteCreateRoleTx = 0x24

	TypeAssumeRoleTx = NameRole + "/assume" // no prefix needed as it is middleware
	TypeCreateRoleTx = NameRole + "/create" // prefix needed for dispatcher
)

func init() {
	sdk.TxMapper.
		RegisterImplementation(AssumeRoleTx{}, TypeAssumeRoleTx, ByteAssumeRoleTx).
		RegisterImplementation(CreateRoleTx{}, TypeCreateRoleTx, ByteCreateRoleTx)
}

// AssumeRoleTx is a layered tx that can wrap your normal tx to give it
// the authority to use a given role.
type AssumeRoleTx struct {
	Role data.Bytes  `json:"role"`
	Tx   sdk.Tx `json:"tx"`
}

// NewAssumeRoleTx creates a new wrapper to add a role to a tx execution
func NewAssumeRoleTx(role []byte, tx sdk.Tx) sdk.Tx {
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
func (tx AssumeRoleTx) Wrap() sdk.Tx {
	return sdk.Tx{tx}
}

// CreateRoleTx is used to construct a new role
//
// TODO: add ability to update signers on a role... but that adds a lot
// more complexity to the permissions
type CreateRoleTx struct {
	Role    data.Bytes       `json:"role"`
	MinSigs uint32           `json:"min_sigs"`
	Signers []sdk.Actor `json:"signers"`
}

// NewCreateRoleTx creates a new role, which we can later use
func NewCreateRoleTx(role []byte, minSigs uint32, signers []sdk.Actor) sdk.Tx {
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
func (tx CreateRoleTx) Wrap() sdk.Tx {
	return sdk.Tx{tx}
}
