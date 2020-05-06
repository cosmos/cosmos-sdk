package types

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/tendermint/tendermint/crypto"
	yaml "gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
)

var (
	_ exported.Account        = (*BaseAccount)(nil)
	_ exported.GenesisAccount = (*BaseAccount)(nil)
	_ exported.GenesisAccount = (*ModuleAccount)(nil)
	_ exported.ModuleAccountI = (*ModuleAccount)(nil)
)

// NewBaseAccount creates a new BaseAccount object
func NewBaseAccount(address sdk.AccAddress, pubKey crypto.PubKey, accountNumber, sequence uint64) *BaseAccount {
	acc := &BaseAccount{
		Address:       address,
		AccountNumber: accountNumber,
		Sequence:      sequence,
	}

	acc.SetPubKey(pubKey)

	return acc
}

// ProtoBaseAccount - a prototype function for BaseAccount
func ProtoBaseAccount() exported.Account {
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
func (acc BaseAccount) GetPubKey() (pk crypto.PubKey) {
	if len(acc.PubKey) == 0 {
		return nil
	}

	amino.MustUnmarshalBinaryBare(acc.PubKey, &pk)
	return pk
}

// SetPubKey - Implements sdk.Account.
func (acc *BaseAccount) SetPubKey(pubKey crypto.PubKey) error {
	if pubKey == nil {
		acc.PubKey = nil
	} else {
		acc.PubKey = pubKey.Bytes()
	}

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
	if len(acc.PubKey) != 0 && acc.Address != nil &&
		!bytes.Equal(acc.GetPubKey().Address().Bytes(), acc.Address.Bytes()) {
		return errors.New("account address and pubkey address do not match")
	}

	return nil
}

func (acc BaseAccount) String() string {
	out, _ := acc.MarshalYAML()
	return out.(string)
}

type baseAccountPretty struct {
	Address       sdk.AccAddress `json:"address" yaml:"address"`
	PubKey        string         `json:"public_key" yaml:"public_key"`
	AccountNumber uint64         `json:"account_number" yaml:"account_number"`
	Sequence      uint64         `json:"sequence" yaml:"sequence"`
}

// MarshalYAML returns the YAML representation of an account.
func (acc BaseAccount) MarshalYAML() (interface{}, error) {
	alias := baseAccountPretty{
		Address:       acc.Address,
		AccountNumber: acc.AccountNumber,
		Sequence:      acc.Sequence,
	}

	if acc.PubKey != nil {
		pks, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, acc.GetPubKey())
		if err != nil {
			return nil, err
		}

		alias.PubKey = pks
	}

	bz, err := yaml.Marshal(alias)
	if err != nil {
		return nil, err
	}

	return string(bz), err
}

// NewModuleAddress creates an AccAddress from the hash of the module's name
func NewModuleAddress(name string) sdk.AccAddress {
	return sdk.AccAddress(crypto.AddressHash([]byte(name)))
}

// NewEmptyModuleAccount creates a empty ModuleAccount from a string
func NewEmptyModuleAccount(name string, permissions ...string) *ModuleAccount {
	moduleAddress := NewModuleAddress(name)
	baseAcc := NewBaseAccountWithAddress(moduleAddress)

	if err := validatePermissions(permissions...); err != nil {
		panic(err)
	}

	return &ModuleAccount{
		BaseAccount: baseAcc,
		Name:        name,
		Permissions: permissions,
	}
}

// NewModuleAccount creates a new ModuleAccount instance
func NewModuleAccount(ba *BaseAccount, name string, permissions ...string) *ModuleAccount {
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

// Validate checks for errors on the account fields
func (ma ModuleAccount) Validate() error {
	if strings.TrimSpace(ma.Name) == "" {
		return errors.New("module account name cannot be blank")
	}

	if !ma.Address.Equals(sdk.AccAddress(crypto.AddressHash([]byte(ma.Name)))) {
		return fmt.Errorf("address %s cannot be derived from the module name '%s'", ma.Address, ma.Name)
	}

	return ma.BaseAccount.Validate()
}

type moduleAccountPretty struct {
	Address       sdk.AccAddress `json:"address" yaml:"address"`
	PubKey        string         `json:"public_key" yaml:"public_key"`
	AccountNumber uint64         `json:"account_number" yaml:"account_number"`
	Sequence      uint64         `json:"sequence" yaml:"sequence"`
	Name          string         `json:"name" yaml:"name"`
	Permissions   []string       `json:"permissions" yaml:"permissions"`
}

func (ma ModuleAccount) String() string {
	out, _ := ma.MarshalYAML()
	return out.(string)
}

// MarshalYAML returns the YAML representation of a ModuleAccount.
func (ma ModuleAccount) MarshalYAML() (interface{}, error) {
	bs, err := yaml.Marshal(moduleAccountPretty{
		Address:       ma.Address,
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

// MarshalJSON returns the JSON representation of a ModuleAccount.
func (ma ModuleAccount) MarshalJSON() ([]byte, error) {
	return json.Marshal(moduleAccountPretty{
		Address:       ma.Address,
		PubKey:        "",
		AccountNumber: ma.AccountNumber,
		Sequence:      ma.Sequence,
		Name:          ma.Name,
		Permissions:   ma.Permissions,
	})
}

// UnmarshalJSON unmarshals raw JSON bytes into a ModuleAccount.
func (ma *ModuleAccount) UnmarshalJSON(bz []byte) error {
	var alias moduleAccountPretty
	if err := json.Unmarshal(bz, &alias); err != nil {
		return err
	}

	ma.BaseAccount = NewBaseAccount(alias.Address, nil, alias.AccountNumber, alias.Sequence)
	ma.Name = alias.Name
	ma.Permissions = alias.Permissions

	return nil
}
