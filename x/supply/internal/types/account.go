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
	Name       string `json:"name"`       // name of the module
	Permission string `json:"permission"` // permission of module account (minter/burner/holder)
}

// NewModuleAddress creates an AccAddress from the hash of the module's name
func NewModuleAddress(name string) sdk.AccAddress {
	return sdk.AccAddress(crypto.AddressHash([]byte(name)))
}

func NewEmptyModuleAccount(name, permission string) *ModuleAccount {
	moduleAddress := NewModuleAddress(name)
	baseAcc := authtypes.NewBaseAccountWithAddress(moduleAddress)

	if err := validatePermissions(permission); err != nil {
		panic(err)
	}

	return &ModuleAccount{
		BaseAccount: &baseAcc,
		Name:        name,
		Permission:  permission,
	}
}

// NewModuleAccount creates a new ModuleAccount instance
func NewModuleAccount(ba *authtypes.BaseAccount,
	name, permission string) *ModuleAccount {

	if err := validatePermissions(permission); err != nil {
		panic(err)
	}

	return &ModuleAccount{
		BaseAccount: ba,
		Name:        name,
		Permission:  permission,
	}
}

// GetName returns the the name of the holder's module
func (ma ModuleAccount) GetName() string {
	return ma.Name
}

// GetPermission returns permission granted to the module account (holder/minter/burner)
func (ma ModuleAccount) GetPermission() string {
	return ma.Permission
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
		Permission    string
	}{
		Address:       ma.Address,
		Coins:         ma.Coins,
		PubKey:        "",
		AccountNumber: ma.AccountNumber,
		Sequence:      ma.Sequence,
		Name:          ma.Name,
		Permission:    ma.Permission,
	})

	if err != nil {
		return nil, err
	}

	return string(bs), nil
}
