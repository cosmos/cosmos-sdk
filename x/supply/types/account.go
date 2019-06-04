package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/tendermint/tendermint/crypto"
)

// ModuleAccount defines an account type for pools that hold tokens in an escrow
type ModuleAccount interface {
	auth.Account

	Name() string
	IsMinter() bool
}

//-----------------------------------------------------------------------------
// Module Holder Account

var _ ModuleAccount = (*ModuleHolderAccount)(nil)

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

// Name returns the the name of the holder's module
func (mha ModuleHolderAccount) Name() string {
	return mha.ModuleName
}

// IsMinter false for ModuleHolderAccount
func (mha ModuleHolderAccount) IsMinter() bool { return false }

// SetPubKey - Implements Account
func (mha *ModuleHolderAccount) SetPubKey(pubKey crypto.PubKey) error {
	return fmt.Errorf("not supported for pool accounts")
}

// SetSequence - Implements Account
func (mha *ModuleHolderAccount) SetSequence(seq uint64) error {
	return fmt.Errorf("not supported for pool accounts")
}

// String follows stringer interface
func (mha ModuleHolderAccount) String() string {
	// we ignore the other fields as they will always be empty
	return fmt.Sprintf(`Module Holder Account:
Name:     			%s
Address:  			%s
Coins:    			%s
Account Number: %d`,
		mha.ModuleName, mha.Address, mha.Coins, mha.AccountNumber)
}

//-----------------------------------------------------------------------------
// Module Minter Account

var _ ModuleAccount = (*ModuleMinterAccount)(nil)

// ModuleMinterAccount defines an account for modules that holds coins on a pool
type ModuleMinterAccount struct {
	*ModuleHolderAccount
}

// NewModuleMinterAccount creates a new  ModuleMinterAccount instance
func NewModuleMinterAccount(name string) *ModuleMinterAccount {
	moduleHolderAcc := NewModuleHolderAccount(name)

	return &ModuleMinterAccount{ModuleHolderAccount: moduleHolderAcc}
}

// IsMinter true for ModuleMinterAccount
func (mma ModuleMinterAccount) IsMinter() bool { return true }

// String follows stringer interface
func (mma ModuleMinterAccount) String() string {
	// we ignore the other fields as they will always be empty
	return fmt.Sprintf(`Module Minter Account:
Name:    				%s
Address: 				%s
Coins:   				%s
Account Number: %d`,
		mma.ModuleName, mma.Address, mma.Coins, mma.AccountNumber)
}
