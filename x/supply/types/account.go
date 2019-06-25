package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/cosmos/cosmos-sdk/x/supply/exported"
	yaml "gopkg.in/yaml.v2"
)

var _ exported.ModuleAccountI = (*ModuleAccount)(nil)

// ModuleAccount defines an account for modules that holds coins on a pool
type ModuleAccount struct {
	*authtypes.BaseAccount
	Name       string `json:"name"` // name of the module
	Permission string `json:"permission"` // permission of module account (minter/burner/holder)
}

// NewModuleAccount creates a new ModuleAccount instance
func NewModuleAccount(name, permission string) *ModuleAccount {
	moduleAddress := sdk.AccAddress(crypto.AddressHash([]byte(name)))
	baseAcc := authtypes.NewBaseAccountWithAddress(moduleAddress)

	if err := validatePermission(permission); err != nil {
		panic(err)
	}

	return &ModuleAccount{
		BaseAccount: &baseAcc,
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
