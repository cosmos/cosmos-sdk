package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/tendermint/tendermint/crypto"
)

// TODO: register within auth codec

// ModuleAccount defines an account type for modules that hold tokens in an escrow
type ModuleAccount interface {
	auth.Account

	Name() string
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
func NewModuleHolderAccount(moduleName string) *ModuleHolderAccount {
	moduleAddress := sdk.AccAddress(crypto.AddressHash([]byte(moduleName)))

	baseAcc := auth.NewBaseAccountWithAddress(moduleAddress)
	return &ModuleHolderAccount{
		BaseAccount: &baseAcc,
		ModuleName:  moduleName,
	}
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
	// we ignore the other fields as they will always be empty
	return fmt.Sprintf(`Module Holder Account:
Address:  %s
Coins:    %s
Name:     %s`,
		mha.Address, mha.Coins, mha.ModuleName)
}

//-----------------------------------------------------------------------------
// Module Minter Account

var _ ModuleAccount = (*ModuleMinterAccount)(nil)

// ModuleMinterAccount defines an account for modules that holds coins on a pool
type ModuleMinterAccount struct {
	*ModuleHolderAccount
}

// NewModuleMinterAccount creates a new  ModuleMinterAccount instance
func NewModuleMinterAccount(moduleName string) *ModuleMinterAccount {
	moduleHolderAcc := NewModuleHolderAccount(moduleName)

	return &ModuleMinterAccount{ModuleHolderAccount: moduleHolderAcc}
}

// String follows stringer interface
func (mma ModuleMinterAccount) String() string {
	// we ignore the other fields as they will always be empty
	return fmt.Sprintf(`Module Minter Account:
Address: %s
Coins:   %s
Name:    %s`,
		mma.Address, mma.Coins, mma.ModuleName)
}
