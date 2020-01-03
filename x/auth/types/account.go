package types

import (
	"bytes"
	"encoding/json"
	"errors"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto"
	tmcrypto "github.com/tendermint/tendermint/crypto"
	yaml "gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
)

//-----------------------------------------------------------------------------
// BaseAccount

var _ exported.Account = (*BaseAccount)(nil)
var _ exported.GenesisAccount = (*BaseAccount)(nil)

// NewBaseAccount creates a new BaseAccount object
func NewBaseAccount(
	address sdk.AccAddress, coins sdk.Coins, pubKey tmcrypto.PubKey, accountNumber, sequence uint64,
) *BaseAccount {

	return &BaseAccount{
		Address:       address,
		Coins:         coins,
		PublicKey:     pubKey.Bytes(),
		AccountNumber: accountNumber,
		Sequence:      sequence,
	}
}

// ProtoBaseAccount - a prototype function for BaseAccount
func ProtoBaseAccount() exported.Account {
	return &BaseAccount{}
}

// NewBaseAccountWithAddress - returns a new base account with a given address
func NewBaseAccountWithAddress(addr sdk.AccAddress) BaseAccount {
	return BaseAccount{
		Address: addr,
	}
}

// GetAddress returns the account's address.
func (m *BaseAccount) GetAddress() sdk.AccAddress {
	if m != nil {
		return m.Address
	}
	return nil
}

// GetCoins returns the account's coins.
func (m *BaseAccount) GetCoins() sdk.Coins {
	if m != nil {
		return m.Coins
	}
	return nil
}

// GetPubKey return's the account's public key.
func (m *BaseAccount) GetPubKey() *crypto.PublicKey {
	if m != nil {
		return m.PublicKey
	}
	return nil
}

// GetAccountNumber returns the account's account number.
func (m *BaseAccount) GetAccountNumber() uint64 {
	if m != nil {
		return m.AccountNumber
	}
	return 0
}

// GetSequence return's the account's sequence (nonce).
func (m *BaseAccount) GetSequence() uint64 {
	if m != nil {
		return m.Sequence
	}
	return 0
}

// SetAddress - Implements sdk.Account.
func (acc *BaseAccount) SetAddress(addr sdk.AccAddress) error {
	if len(acc.Address) != 0 {
		return errors.New("cannot override BaseAccount address")
	}
	acc.Address = addr
	return nil
}

// SetPubKey - Implements sdk.Account.
func (acc *BaseAccount) SetPubKey(pubKey tmcrypto.PubKey) error {
	acc.PubKey = pubKey
	return nil
}

// SetCoins - Implements sdk.Account.
func (acc *BaseAccount) SetCoins(coins sdk.Coins) error {
	acc.Coins = coins
	return nil
}

// SetAccountNumber - Implements Account
func (acc *BaseAccount) SetAccountNumber(accNumber uint64) error {
	acc.AccountNumber = accNumber
	return nil
}

// SetSequence - Implements sdk.Account.
func (acc *BaseAccount) SetSequence(seq uint64) error {
	acc.Sequence = seq
	return nil
}

// SpendableCoins returns the total set of spendable coins. For a base account,
// this is simply the base coins.
func (acc *BaseAccount) SpendableCoins(_ time.Time) sdk.Coins {
	return acc.GetCoins()
}

// Validate checks for errors on the account fields
func (acc BaseAccount) Validate() error {
	if acc.PubKey != nil && acc.Address != nil &&
		!bytes.Equal(acc.PubKey.Address().Bytes(), acc.Address.Bytes()) {
		return errors.New("pubkey and address pair is invalid")
	}

	return nil
}

type baseAccountPretty struct {
	Address       sdk.AccAddress `json:"address" yaml:"address"`
	Coins         sdk.Coins      `json:"coins" yaml:"coins"`
	PubKey        string         `json:"public_key" yaml:"public_key"`
	AccountNumber uint64         `json:"account_number" yaml:"account_number"`
	Sequence      uint64         `json:"sequence" yaml:"sequence"`
}

func (acc BaseAccount) String() string {
	out, _ := acc.MarshalYAML()
	return out.(string)
}

// MarshalYAML returns the YAML representation of an account.
func (acc BaseAccount) MarshalYAML() (interface{}, error) {
	alias := baseAccountPretty{
		Address:       acc.Address,
		Coins:         acc.Coins,
		AccountNumber: acc.AccountNumber,
		Sequence:      acc.Sequence,
	}

	if acc.PubKey != nil {
		pks, err := sdk.Bech32ifyAccPub(acc.PubKey)
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
		Coins:         acc.Coins,
		AccountNumber: acc.AccountNumber,
		Sequence:      acc.Sequence,
	}

	if acc.PubKey != nil {
		pks, err := sdk.Bech32ifyAccPub(acc.PubKey)
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

	if alias.PubKey != "" {
		pk, err := sdk.GetAccPubKeyBech32(alias.PubKey)
		if err != nil {
			return err
		}

		acc.PubKey = pk
	}

	acc.Address = alias.Address
	acc.Coins = alias.Coins
	acc.AccountNumber = alias.AccountNumber
	acc.Sequence = alias.Sequence

	return nil
}
