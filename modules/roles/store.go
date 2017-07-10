package roles

import (
	"fmt"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/state"
	wire "github.com/tendermint/go-wire"
)

// Role - structure to hold permissioning
type Role struct {
	MinSigs uint32           `json:"min_sigs"`
	Signers []basecoin.Actor `json:"signers"`
}

// MakeKey creates the lookup key for a role
func MakeKey(role []byte) []byte {
	prefix := []byte(NameRole + "/")
	return append(prefix, role...)
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

func createRole(store state.KVStore, key []byte, role Role) error {
	if _, err := loadRole(store, key); !IsNoRoleErr(err) {
		return ErrRoleExists()
	}
	bin := wire.BinaryBytes(role)
	store.Set(key, bin)
	return nil // real stores can return error...
}
