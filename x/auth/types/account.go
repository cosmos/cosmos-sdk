package types

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/tendermint/tendermint/crypto"
	yaml "gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
)

//-----------------------------------------------------------------------------
// BaseAccount

var _ exported.AccountI = (*BaseAccount)(nil)
var _ exported.GenesisAccount = (*BaseAccount)(nil)

// NewBaseAccount creates a new BaseAccount object
func NewBaseAccount(address sdk.AccAddress, pubKey crypto.PubKey, accountNumber, sequence uint64) *BaseAccount {
	var pkStr string
	if pubKey != nil {
		pkStr = sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, pubKey)
	}

	return &BaseAccount{
		Address:       address,
		PubKey:        pkStr,
		AccountNumber: accountNumber,
		Sequence:      sequence,
	}
}

// ProtoBaseAccount - a prototype function for BaseAccount
func ProtoBaseAccount() exported.AccountI {
	return &BaseAccount{}
}

// NewBaseAccountWithAddress - returns a new base account with a given address
func NewBaseAccountWithAddress(addr sdk.AccAddress) *BaseAccount {
	return &BaseAccount{
		Address: addr,
	}
}

// GetAddress - Implements sdk.Account.
func (acc BaseAccount) GetAddress() sdk.AccAddress {
	return acc.Address
}

// SetAddress - Implements sdk.Account.
func (acc *BaseAccount) SetAddress(addr sdk.AccAddress) error {
	if len(acc.Address) != 0 {
		return errors.New("cannot override BaseAccount address")
	}

	acc.Address = addr
	return nil
}

// GetPubKey - Implements sdk.Account.
func (acc BaseAccount) GetPubKey() crypto.PubKey {
	if acc.PubKey == "" {
		return nil
	}

	return sdk.MustGetPubKeyFromBech32(sdk.Bech32PubKeyTypeAccPub, acc.PubKey)
}

// SetPubKey - Implements sdk.Account.
func (acc *BaseAccount) SetPubKey(pubKey crypto.PubKey) error {
	var (
		pkStr string
		err   error
	)

	if pubKey != nil {
		pkStr, err = sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, pubKey)
		if err != nil {
			return err
		}
	}

	acc.PubKey = pkStr
	return nil
}

// GetAccountNumber - Implements Account
func (acc BaseAccount) GetAccountNumber() uint64 {
	return acc.AccountNumber
}

// SetAccountNumber - Implements Account
func (acc *BaseAccount) SetAccountNumber(accNumber uint64) error {
	acc.AccountNumber = accNumber
	return nil
}

// GetSequence - Implements sdk.Account.
func (acc BaseAccount) GetSequence() uint64 {
	return acc.Sequence
}

// SetSequence - Implements sdk.Account.
func (acc *BaseAccount) SetSequence(seq uint64) error {
	acc.Sequence = seq
	return nil
}

// Validate checks for errors on the account fields
func (acc BaseAccount) Validate() error {
	if acc.PubKey != "" && acc.Address != nil &&
		!bytes.Equal(acc.GetPubKey().Address().Bytes(), acc.Address.Bytes()) {
		return errors.New("pubkey and address pair is invalid")
	}

	return nil
}

func (acc BaseAccount) String() string {
	out, _ := yaml.Marshal(acc)
	return string(out)
}

// SetAccountI sets the Account's oneof sum type to the provided AccountI type.
// The provided AccountI type must be a reference to a BaseAccount.
func (acc *Account) SetAccountI(value exported.AccountI) error {
	if value == nil {
		acc.Sum = nil
		return nil
	}

	baseAcc, ok := value.(*BaseAccount)
	if ok {
		acc.Sum = &Account_BaseAccount{baseAcc}
		return nil
	}

	return fmt.Errorf("failed to encode value of type %T as message AccountI", value)
}

// GetAccountI returns an AccountI based on the internal oneof sum type.
func (acc *Account) GetAccountI() exported.AccountI {
	if x := acc.GetBaseAccount(); x != nil {
		return x
	}

	return nil
}
