package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/tendermint/tendermint/crypto"
	yaml "gopkg.in/yaml.v2"
)

var _ ModuleAccountI = (*ModuleAccount)(nil)

// ModuleAccount defines an account type for pools that hold tokens in an escrow
type ModuleAccountI interface {
	auth.Account
	GetName() string
	GetPermission() string
}

// ModuleAccount defines an account for modules that holds coins on a pool
type ModuleAccount struct {
	*auth.BaseAccount
	Name       string `json:"name"` // name of the module
	Permission string `json:"perm"` // permission of module account (minter/burner/holder)
}

// NewModuleAccount creates a new ModuleAccount instance
func NewModuleAccount(name, permission string) *ModuleAccount {
	moduleAddress := sdk.AccAddress(crypto.AddressHash([]byte(name)))
	baseAcc := auth.NewBaseAccountWithAddress(moduleAddress)

	if err := validatePermission(permission); err != nil {
		panic(err)
	}

	return &ModuleAccount{
		BaseAccount: &baseAcc,
		Name:        name,
		Permission:  permission,
	}
}

// Name returns the the name of the holder's module
func (ma ModuleAccount) GetName() string {
	return ma.Name
}

// Name returns the the name of the holder's module
func (ma ModuleAccount) GetPermission() string {
	return ma.Permission
}

// SetPubKey - Implements Account
func (ma *ModuleAccount) SetPubKey(pubKey crypto.PubKey) error {
	return fmt.Errorf("not supported for module accounts")
}

// SetSequence - Implements Account
func (ma *ModuleAccount) SetSequence(seq uint64) error {
	return fmt.Errorf("not supported for module accounts")
}

// String follows stringer interface
func (ma ModuleAccount) String() string {
	b, err := yaml.Marshal(ma)
	if err != nil {
		panic(err)
	}
	return string(b)
}
