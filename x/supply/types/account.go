package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/tendermint/tendermint/crypto"
	yaml "gopkg.in/yaml.v2"
)

var (
	_ ModuleAccount = (*ModuleHolderAccount)(nil)
	_ ModuleAccount = (*ModuleMinterAccount)(nil)
	_ ModuleAccount = (*ModuleBurnerAccount)(nil)
)

// ModuleAccount defines an account type for pools that hold tokens in an escrow
type ModuleAccount interface {
	auth.Account
	Name() string
}

// _____________________________________________________________________
// Module Holder Account

// ModuleHolderAccount defines an account for modules that holds coins on a pool
type ModuleHolderAccount struct {
	*auth.BaseAccount
	ModuleName string `json:"name"` // name of the module
}

// NewModuleHolderAccount creates a new ModuleHolderAccount instance
func NewModuleHolderAccount(name string) *ModuleHolderAccount {
	moduleAddress := sdk.AccAddress(crypto.AddressHash([]byte(name)))

	baseAcc := auth.NewBaseAccountWithAddress(moduleAddress)
	return &ModuleHolderAccount{
		BaseAccount: &baseAcc,
		ModuleName:  name,
	}
}

// InitNewModuleHolderAccount initializes and creates a new module holder account
func InitNewModuleHolderAccount(ctx sdk.Context, name string, accKeeper AccountKeeper) *ModuleHolderAccount {
	acc := NewModuleHolderAccount(name)
	if err := acc.SetAccountNumber(accKeeper.GetNextAccountNumber(ctx)); err != nil {
		panic(err)
	}
	accKeeper.SetAccount(ctx, acc)
	return acc
}

// Name returns the the name of the holder's module
func (mha ModuleHolderAccount) Name() string {
	return mha.ModuleName
}

// SetPubKey - Implements Account
func (mha *ModuleHolderAccount) SetPubKey(pubKey crypto.PubKey) error {
	return fmt.Errorf("not supported for module accounts")
}

// SetSequence - Implements Account
func (mha *ModuleHolderAccount) SetSequence(seq uint64) error {
	return fmt.Errorf("not supported for module accounts")
}

// String follows stringer interface
func (mha ModuleHolderAccount) String() string {
	b, err := yaml.Marshal(mha)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// ___________________________________________________________________

// ModuleMinterAccount defines an account for modules that can hold and mint tokens
type ModuleMinterAccount struct {
	*ModuleHolderAccount
}

// NewModuleMinterAccount creates a new  ModuleMinterAccount instance
func NewModuleMinterAccount(name string) *ModuleMinterAccount {
	acc := NewModuleHolderAccount(name)

	return &ModuleMinterAccount{ModuleHolderAccount: acc}
}

// InitNewModuleMinterAccount initializes and creates a new module minter account
func InitNewModuleMinterAccount(ctx sdk.Context, name string, accKeeper AccountKeeper) *ModuleMinterAccount {
	acc := NewModuleMinterAccount(name)
	if err := acc.SetAccountNumber(accKeeper.GetNextAccountNumber(ctx)); err != nil {
		panic(err)
	}
	accKeeper.SetAccount(ctx, acc)
	return acc
}

// String follows stringer interface
func (mma ModuleMinterAccount) String() string {
	b, err := yaml.Marshal(mma)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// ___________________________________________________________________

// ModuleBurnerAccount defines an account for modules that can hold and mint tokens
type ModuleBurnerAccount struct {
	*ModuleHolderAccount
}

// NewModuleBurnerAccount creates a new  ModuleBurnerAccount instance
func NewModuleBurnerAccount(name string) *ModuleBurnerAccount {
	acc := NewModuleHolderAccount(name)

	return &ModuleBurnerAccount{ModuleHolderAccount: acc}
}

// InitNewModuleBurnerAccount initializes and creates a new module burner account
func InitNewModuleBurnerAccount(ctx sdk.Context, name string, accKeeper AccountKeeper) *ModuleBurnerAccount {
	acc := NewModuleBurnerAccount(name)
	if err := acc.SetAccountNumber(accKeeper.GetNextAccountNumber(ctx)); err != nil {
		panic(err)
	}
	accKeeper.SetAccount(ctx, acc)
	return acc
}

// String follows stringer interface
func (mba ModuleBurnerAccount) String() string {
	b, err := yaml.Marshal(mba)
	if err != nil {
		panic(err)
	}
	return string(b)
}
