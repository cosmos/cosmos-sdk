package types

import (
	"fmt"

	yaml "gopkg.in/yaml.v2"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/supply/exported"
)

var _ exported.ModuleAccountI = (*ModuleAccount)(nil)

// ModuleAccount defines an account for modules that holds coins on a pool
type ModuleAccount struct {
	*authtypes.BaseAccount
	Name        string   `json:"name" yaml:"name"`               // name of the module
	Permissions []string `json:"permissions" yaml:"permissions"` // permissions of module account
}

// NewModuleAddress creates an AccAddress from the hash of the module's name
func NewModuleAddress(name string) sdk.AccAddress {
	return sdk.AccAddress(crypto.AddressHash([]byte(name)))
}

func NewEmptyModuleAccount(name string, permissions ...string) *ModuleAccount {
	moduleAddress := NewModuleAddress(name)
	baseAcc := authtypes.NewBaseAccountWithAddress(moduleAddress)

	if err := validatePermissions(permissions...); err != nil {
		panic(err)
	}

	return &ModuleAccount{
		BaseAccount: &baseAcc,
		Name:        name,
		Permissions: permissions,
	}
}

// NewModuleAccount creates a new ModuleAccount instance
func NewModuleAccount(ba *authtypes.BaseAccount,
	name string, permissions ...string) *ModuleAccount {

	if err := validatePermissions(permissions...); err != nil {
		panic(err)
	}

	return &ModuleAccount{
		BaseAccount: ba,
		Name:        name,
		Permissions: permissions,
	}
}

// HasPermission returns whether or not the module account has permission.
func (ma ModuleAccount) HasPermission(permission string) bool {
	for _, perm := range ma.Permissions {
		if perm == permission {
			return true
		}
	}
	return false
}

// GetName returns the the name of the holder's module
func (ma ModuleAccount) GetName() string {
	return ma.Name
}

// GetPermissions returns permissions granted to the module account
func (ma ModuleAccount) GetPermissions() []string {
	return ma.Permissions
}

// SetPubKey - Implements Account
func (ma ModuleAccount) SetPubKey(pubKey crypto.PubKey) error {
	return fmt.Errorf("not supported for module accounts")
}

// SetSequence - Implements Account
func (ma ModuleAccount) SetSequence(seq uint64) error {
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

// MarshalYAML returns the YAML representation of a ModuleAccount.
func (ma ModuleAccount) MarshalYAML() (interface{}, error) {
	bs, err := yaml.Marshal(struct {
		Address       sdk.AccAddress
		Coins         sdk.Coins
		PubKey        string
		AccountNumber uint64
		Sequence      uint64
		Name          string
		Permissions   []string
	}{
		Address:       ma.Address,
		Coins:         ma.Coins,
		PubKey:        "",
		AccountNumber: ma.AccountNumber,
		Sequence:      ma.Sequence,
		Name:          ma.Name,
		Permissions:   ma.Permissions,
	})

	if err != nil {
		return nil, err
	}

	return string(bs), nil
}
