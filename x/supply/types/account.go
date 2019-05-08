package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/tendermint/tendermint/crypto"
)

// PoolAccount defines an account type for pools that hold tokens in an escrow
type PoolAccount interface {
	auth.Account

	Name() string
}

//-----------------------------------------------------------------------------
// Module Holder Account

var _ PoolAccount = (*PoolHolderAccount)(nil)

// PoolHolderAccount defines an account for modules that holds coins on a pool
type PoolHolderAccount struct {
	*auth.BaseAccount

	PoolName string `json:"name"` // name of the pool
}

// NewPoolHolderAccount creates a new PoolHolderAccount instance
func NewPoolHolderAccount(name string) *PoolHolderAccount {
	moduleAddress := sdk.AccAddress(crypto.AddressHash([]byte(name)))

	baseAcc := auth.NewBaseAccountWithAddress(moduleAddress)
	return &PoolHolderAccount{
		BaseAccount: &baseAcc,
		PoolName:    name,
	}
}

// Name returns the the name of the holder's module
func (pha PoolHolderAccount) Name() string {
	return pha.PoolName
}

// SetPubKey - Implements Account
func (pha *PoolHolderAccount) SetPubKey(pubKey crypto.PubKey) error {
	return fmt.Errorf("not supported for pool accounts")
}

// SetSequence - Implements Account
func (pha *PoolHolderAccount) SetSequence(seq uint64) error {
	return fmt.Errorf("not supported for pool accounts")
}

// String follows stringer interface
func (pha PoolHolderAccount) String() string {
	// we ignore the other fields as they will always be empty
	return fmt.Sprintf(`Pool Holder Account:
Address:  %s
Coins:    %s
Name:     %s`,
		pha.Address, pha.Coins, pha.PoolName)
}

//-----------------------------------------------------------------------------
// Module Minter Account

var _ PoolAccount = (*PoolMinterAccount)(nil)

// PoolMinterAccount defines an account for modules that holds coins on a pool
type PoolMinterAccount struct {
	*PoolHolderAccount
}

// NewPoolMinterAccount creates a new  PoolMinterAccount instance
func NewPoolMinterAccount(name string) *PoolMinterAccount {
	moduleHolderAcc := NewPoolHolderAccount(name)

	return &PoolMinterAccount{PoolHolderAccount: moduleHolderAcc}
}

// String follows stringer interface
func (pma PoolMinterAccount) String() string {
	// we ignore the other fields as they will always be empty
	return fmt.Sprintf(`Pool Minter Account:
Address: %s
Coins:   %s
Name:    %s`,
		pma.Address, pma.Coins, pma.PoolName)
}
