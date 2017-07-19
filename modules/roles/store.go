package roles

import (
	"fmt"

	wire "github.com/tendermint/go-wire"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/state"
)

// NewPerm creates a role permission with the given label
func NewPerm(role []byte) basecoin.Actor {
	return basecoin.Actor{
		App:     NameRole,
		Address: role,
	}
}

// Role - structure to hold permissioning
type Role struct {
	MinSigs uint32           `json:"min_sigs"`
	Signers []basecoin.Actor `json:"signers"`
}

// NewRole creates a Role structure to store the permissioning
func NewRole(min uint32, signers []basecoin.Actor) Role {
	return Role{
		MinSigs: min,
		Signers: signers,
	}
}

// IsSigner checks if the given Actor is allowed to sign this role
func (r Role) IsSigner(a basecoin.Actor) bool {
	for _, s := range r.Signers {
		if a.Equals(s) {
			return true
		}
	}
	return false
}

// IsAuthorized checks if the context has permission to assume the role
func (r Role) IsAuthorized(ctx basecoin.Context) bool {
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

func loadRole(store state.KVStore, key []byte) (role Role, err error) {
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

func checkNoRole(store state.KVStore, key []byte) error {
	if _, err := loadRole(store, key); !IsNoRoleErr(err) {
		return ErrRoleExists()
	}
	return nil
}

// we only have create here, no update, since we don't allow update yet
func createRole(store state.KVStore, key []byte, role Role) error {
	if err := checkNoRole(store, key); err != nil {
		return err
	}
	bin := wire.BinaryBytes(role)
	store.Set(key, bin)
	return nil // real stores can return error...
}
