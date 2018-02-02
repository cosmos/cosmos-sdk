package roles

import (
	"fmt"

	wire "github.com/tendermint/go-wire"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/errors"
	"github.com/cosmos/cosmos-sdk/state"
)

// NewPerm creates a role permission with the given label
func NewPerm(role []byte) sdk.Actor {
	return sdk.Actor{
		App:     NameRole,
		Address: role,
	}
}

// Role - structure to hold permissioning
type Role struct {
	MinSigs uint32           `json:"min_sigs"`
	Signers []sdk.Actor `json:"signers"`
}

// NewRole creates a Role structure to store the permissioning
func NewRole(min uint32, signers []sdk.Actor) Role {
	return Role{
		MinSigs: min,
		Signers: signers,
	}
}

// IsSigner checks if the given Actor is allowed to sign this role
func (r Role) IsSigner(a sdk.Actor) bool {
	for _, s := range r.Signers {
		if a.Equals(s) {
			return true
		}
	}
	return false
}

// IsAuthorized checks if the context has permission to assume the role
func (r Role) IsAuthorized(ctx sdk.Context) bool {
	needed := r.MinSigs
	for _, s := range r.Signers {
		if ctx.HasPermission(s) {
			needed--
			if needed <= 0 {
				return true
			}
		}
	}
	return false
}

func loadRole(store state.SimpleDB, key []byte) (role Role, err error) {
	data := store.Get(key)
	if len(data) == 0 {
		return role, ErrNoRole()
	}
	err = wire.ReadBinaryBytes(data, &role)
	if err != nil {
		msg := fmt.Sprintf("Error reading role %X", key)
		return role, errors.ErrInternal(msg)
	}
	return role, nil
}

func checkNoRole(store state.SimpleDB, key []byte) error {
	if _, err := loadRole(store, key); !IsNoRoleErr(err) {
		return ErrRoleExists()
	}
	return nil
}

// we only have create here, no update, since we don't allow update yet
func createRole(store state.SimpleDB, key []byte, role Role) error {
	if err := checkNoRole(store, key); err != nil {
		return err
	}
	bin := wire.BinaryBytes(role)
	store.Set(key, bin)
	return nil // real stores can return error...
}
