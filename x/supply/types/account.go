package types

import (
	"errors"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/tendermint/tendermint/crypto"
)

// PoolAccount defines an account type for pools that hold tokens in an escrow
type PoolAccount interface {
	auth.Account

	Name() string
	GetDecCoins() sdk.DecCoins
	SetDecCoins(sdk.DecCoins) error
}

//-----------------------------------------------------------------------------
// Module Holder Account

var _ PoolAccount = (*PoolHolderAccount)(nil)

// PoolHolderAccount defines an account for modules that holds coins on a pool
type PoolHolderAccount struct {
	Address       sdk.AccAddress `json:"address"`
	Coins         sdk.DecCoins   `json:"coins"`
	PoolName      string         `json:"name"` // name of the pool
	AccountNumber uint64         `json:"account_number"`
}

// NewPoolHolderAccount creates a new PoolHolderAccount instance
func NewPoolHolderAccount(name string) *PoolHolderAccount {
	poolAddress := sdk.AccAddress(crypto.AddressHash([]byte(name)))

	return &PoolHolderAccount{
		Address:  poolAddress,
		PoolName: name,
	}
}

// Name returns the the name of the holder's module
func (pha PoolHolderAccount) Name() string {
	return pha.PoolName
}

// GetCoins - Implements Account
func (pha PoolHolderAccount) GetCoins() sdk.Coins {
	coins, _ := pha.Coins.TruncateDecimal()
	return coins
}

// SetCoins - Implements Account
func (pha *PoolHolderAccount) SetCoins(coins sdk.Coins) error {
	pha.Coins = sdk.NewDecCoins(coins)
	return nil
}

// GetDecCoins - returns the account decimal coins
func (pha PoolHolderAccount) GetDecCoins() sdk.DecCoins {
	return pha.Coins
}

// SetDecCoins - sets decimal coins
func (pha *PoolHolderAccount) SetDecCoins(coins sdk.DecCoins) error {
	pha.Coins = coins
	return nil
}

// GetAddress - Implements sdk.Account.
func (pha PoolHolderAccount) GetAddress() sdk.AccAddress {
	return pha.Address
}

// SetAddress - Implements sdk.Account.
func (pha *PoolHolderAccount) SetAddress(addr sdk.AccAddress) error {
	if len(pha.Address) != 0 {
		return errors.New("cannot override PoolHolderAccount address")
	}
	pha.Address = addr
	return nil
}

// GetPubKey - Implements Account
func (pha PoolHolderAccount) GetPubKey() crypto.PubKey {
	return nil
}

// SetPubKey - Implements Account
func (pha *PoolHolderAccount) SetPubKey(pubKey crypto.PubKey) error {
	return fmt.Errorf("not supported for pool accounts")
}

// GetAccountNumber - Implements Account
func (pha PoolHolderAccount) GetAccountNumber() uint64 {
	return pha.AccountNumber
}

// SetAccountNumber - Implements Account
func (pha *PoolHolderAccount) SetAccountNumber(accNumber uint64) error {
	pha.AccountNumber = accNumber
	return nil
}

// GetSequence - Implements Account
func (pha PoolHolderAccount) GetSequence() uint64 {
	return uint64(0)
}

// SpendableCoins - Implements Account
func (pha PoolHolderAccount) SpendableCoins(blockTime time.Time) sdk.Coins {
	return pha.GetCoins()
}

// SetSequence - Implements Account
func (pha *PoolHolderAccount) SetSequence(seq uint64) error {
	return fmt.Errorf("not supported for pool accounts")
}

// String follows stringer interface
func (pha PoolHolderAccount) String() string {
	// we ignore the other fields as they will always be empty
	return fmt.Sprintf(`Pool Holder Account:
Address:        %s
Coins:          %s
Name:           %s
Account Number: %d`,
		pha.Address, pha.Coins, pha.PoolName, pha.AccountNumber)
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
	Address:        %s
	Coins:          %s
	Name:           %s
	Account Number: %d`,
		pma.Address, pma.Coins, pma.PoolName, pma.AccountNumber)
}
