package types

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/tendermint/tendermint/crypto"
	yaml "gopkg.in/yaml.v2"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
)

var _ exported.Account = (*BaseAccount)(nil)
var _ exported.GenesisAccount = (*BaseAccount)(nil)

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

	codec.Cdc.MustUnmarshalBinaryBare(acc.PubKey, &pk)
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

// MarshalJSON returns the JSON representation of a BaseAccount.
func (acc BaseAccount) MarshalJSON() ([]byte, error) {
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

	return json.Marshal(alias)
}

// UnmarshalJSON unmarshals raw JSON bytes into a BaseAccount.
func (acc *BaseAccount) UnmarshalJSON(bz []byte) error {
	var alias baseAccountPretty
	if err := json.Unmarshal(bz, &alias); err != nil {
		return err
	}

	// NOTE: This will not work for multisig-based accounts as their Bech32
	// encoding is too long.
	if alias.PubKey != "" {
		pk, err := sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeAccPub, alias.PubKey)
		if err != nil {
			return err
		}

		acc.PubKey = pk.Bytes()
	}

	acc.Address = alias.Address
	acc.AccountNumber = alias.AccountNumber
	acc.Sequence = alias.Sequence

	return nil
}
