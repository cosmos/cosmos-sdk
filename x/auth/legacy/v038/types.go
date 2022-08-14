// Package v038 is used for legacy migration scripts. Actual migration scripts
// for v038 have been removed, but the v039->v042 migration script still
// references types from this file, so we're keeping it for now.
package v038

// DONTCOVER

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	tmcrypto "github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32/legacybech32"
	v034auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v034"
)

const (
	ModuleName = "auth"
)

type (
	// partial interface needed only for amino encoding and sanitization
	Account interface {
		GetAddress() sdk.AccAddress
		GetAccountNumber() uint64
		GetCoins() sdk.Coins
		SetCoins(sdk.Coins) error
	}

	GenesisAccount interface {
		Account

		Validate() error
	}

	GenesisAccounts []GenesisAccount

	GenesisState struct {
		Params   v034auth.Params `json:"params" yaml:"params"`
		Accounts GenesisAccounts `json:"accounts" yaml:"accounts"`
	}

	BaseAccount struct {
		Address       sdk.AccAddress     `json:"address" yaml:"address"`
		Coins         sdk.Coins          `json:"coins,omitempty" yaml:"coins,omitempty"`
		PubKey        cryptotypes.PubKey `json:"public_key" yaml:"public_key"`
		AccountNumber uint64             `json:"account_number" yaml:"account_number"`
		Sequence      uint64             `json:"sequence" yaml:"sequence"`
	}

	baseAccountPretty struct {
		Address       sdk.AccAddress `json:"address" yaml:"address"`
		Coins         sdk.Coins      `json:"coins,omitempty" yaml:"coins,omitempty"`
		PubKey        string         `json:"public_key" yaml:"public_key"`
		AccountNumber uint64         `json:"account_number" yaml:"account_number"`
		Sequence      uint64         `json:"sequence" yaml:"sequence"`
	}

	BaseVestingAccount struct {
		*BaseAccount

		OriginalVesting  sdk.Coins `json:"original_vesting"`
		DelegatedFree    sdk.Coins `json:"delegated_free"`
		DelegatedVesting sdk.Coins `json:"delegated_vesting"`

		EndTime int64 `json:"end_time"`
	}

	vestingAccountPretty struct {
		Address          sdk.AccAddress `json:"address" yaml:"address"`
		Coins            sdk.Coins      `json:"coins,omitempty" yaml:"coins,omitempty"`
		PubKey           string         `json:"public_key" yaml:"public_key"`
		AccountNumber    uint64         `json:"account_number" yaml:"account_number"`
		Sequence         uint64         `json:"sequence" yaml:"sequence"`
		OriginalVesting  sdk.Coins      `json:"original_vesting" yaml:"original_vesting"`
		DelegatedFree    sdk.Coins      `json:"delegated_free" yaml:"delegated_free"`
		DelegatedVesting sdk.Coins      `json:"delegated_vesting" yaml:"delegated_vesting"`
		EndTime          int64          `json:"end_time" yaml:"end_time"`

		// custom fields based on concrete vesting type which can be omitted
		StartTime int64 `json:"start_time,omitempty" yaml:"start_time,omitempty"`
	}

	ContinuousVestingAccount struct {
		*BaseVestingAccount

		StartTime int64 `json:"start_time"`
	}

	DelayedVestingAccount struct {
		*BaseVestingAccount
	}

	ModuleAccount struct {
		*BaseAccount

		Name        string   `json:"name" yaml:"name"`
		Permissions []string `json:"permissions" yaml:"permissions"`
	}

	moduleAccountPretty struct {
		Address       sdk.AccAddress `json:"address" yaml:"address"`
		Coins         sdk.Coins      `json:"coins,omitempty" yaml:"coins,omitempty"`
		PubKey        string         `json:"public_key" yaml:"public_key"`
		AccountNumber uint64         `json:"account_number" yaml:"account_number"`
		Sequence      uint64         `json:"sequence" yaml:"sequence"`
		Name          string         `json:"name" yaml:"name"`
		Permissions   []string       `json:"permissions" yaml:"permissions"`
	}
)

func NewGenesisState(params v034auth.Params, accounts GenesisAccounts) GenesisState {
	return GenesisState{
		Params:   params,
		Accounts: accounts,
	}
}

func NewBaseAccountWithAddress(addr sdk.AccAddress) BaseAccount {
	return BaseAccount{
		Address: addr,
	}
}

func NewBaseAccount(
	address sdk.AccAddress, coins sdk.Coins, pk cryptotypes.PubKey, accountNumber, sequence uint64,
) *BaseAccount {

	return &BaseAccount{
		Address:       address,
		Coins:         coins,
		PubKey:        pk,
		AccountNumber: accountNumber,
		Sequence:      sequence,
	}
}

func (acc BaseAccount) GetAddress() sdk.AccAddress {
	return acc.Address
}

func (acc *BaseAccount) GetAccountNumber() uint64 {
	return acc.AccountNumber
}

func (acc *BaseAccount) GetCoins() sdk.Coins {
	return acc.Coins
}

func (acc *BaseAccount) SetCoins(coins sdk.Coins) error {
	acc.Coins = coins
	return nil
}

func (acc BaseAccount) Validate() error {
	if acc.PubKey != nil && acc.Address != nil &&
		!bytes.Equal(acc.PubKey.Address().Bytes(), acc.Address.Bytes()) {
		return errors.New("pubkey and address pair is invalid")
	}

	return nil
}

func (acc BaseAccount) MarshalJSON() ([]byte, error) {
	alias := baseAccountPretty{
		Address:       acc.Address,
		Coins:         acc.Coins,
		AccountNumber: acc.AccountNumber,
		Sequence:      acc.Sequence,
	}

	if acc.PubKey != nil {
		pks, err := legacybech32.MarshalPubKey(legacybech32.AccPK, acc.PubKey)
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
		pk, err := legacybech32.UnmarshalPubKey(legacybech32.AccPK, alias.PubKey)
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

func NewBaseVestingAccount(
	baseAccount *BaseAccount, originalVesting, delegatedFree, delegatedVesting sdk.Coins, endTime int64,
) *BaseVestingAccount {

	return &BaseVestingAccount{
		BaseAccount:      baseAccount,
		OriginalVesting:  originalVesting,
		DelegatedFree:    delegatedFree,
		DelegatedVesting: delegatedVesting,
		EndTime:          endTime,
	}
}

func (bva BaseVestingAccount) Validate() error {
	return bva.BaseAccount.Validate()
}

// MarshalJSON returns the JSON representation of a BaseVestingAccount.
func (bva BaseVestingAccount) MarshalJSON() ([]byte, error) {
	alias := vestingAccountPretty{
		Address:          bva.Address,
		Coins:            bva.Coins,
		AccountNumber:    bva.AccountNumber,
		Sequence:         bva.Sequence,
		OriginalVesting:  bva.OriginalVesting,
		DelegatedFree:    bva.DelegatedFree,
		DelegatedVesting: bva.DelegatedVesting,
		EndTime:          bva.EndTime,
	}

	if bva.PubKey != nil {
		pks, err := legacybech32.MarshalPubKey(legacybech32.AccPK, bva.PubKey)
		if err != nil {
			return nil, err
		}

		alias.PubKey = pks
	}

	return json.Marshal(alias)
}

// UnmarshalJSON unmarshals raw JSON bytes into a BaseVestingAccount.
func (bva *BaseVestingAccount) UnmarshalJSON(bz []byte) error {
	var alias vestingAccountPretty
	if err := json.Unmarshal(bz, &alias); err != nil {
		return err
	}

	var (
		pk  cryptotypes.PubKey
		err error
	)

	if alias.PubKey != "" {
		pk, err = legacybech32.UnmarshalPubKey(legacybech32.AccPK, alias.PubKey)
		if err != nil {
			return err
		}
	}

	bva.BaseAccount = NewBaseAccount(alias.Address, alias.Coins, pk, alias.AccountNumber, alias.Sequence)
	bva.OriginalVesting = alias.OriginalVesting
	bva.DelegatedFree = alias.DelegatedFree
	bva.DelegatedVesting = alias.DelegatedVesting
	bva.EndTime = alias.EndTime

	return nil
}

func NewContinuousVestingAccountRaw(bva *BaseVestingAccount, startTime int64) *ContinuousVestingAccount {
	return &ContinuousVestingAccount{
		BaseVestingAccount: bva,
		StartTime:          startTime,
	}
}

func (cva ContinuousVestingAccount) Validate() error {
	if cva.StartTime >= cva.EndTime {
		return errors.New("vesting start-time cannot be before end-time")
	}

	return cva.BaseVestingAccount.Validate()
}

// MarshalJSON returns the JSON representation of a ContinuousVestingAccount.
func (cva ContinuousVestingAccount) MarshalJSON() ([]byte, error) {
	alias := vestingAccountPretty{
		Address:          cva.Address,
		Coins:            cva.Coins,
		AccountNumber:    cva.AccountNumber,
		Sequence:         cva.Sequence,
		OriginalVesting:  cva.OriginalVesting,
		DelegatedFree:    cva.DelegatedFree,
		DelegatedVesting: cva.DelegatedVesting,
		EndTime:          cva.EndTime,
		StartTime:        cva.StartTime,
	}

	if cva.PubKey != nil {
		pks, err := legacybech32.MarshalPubKey(legacybech32.AccPK, cva.PubKey)
		if err != nil {
			return nil, err
		}

		alias.PubKey = pks
	}

	return json.Marshal(alias)
}

// UnmarshalJSON unmarshals raw JSON bytes into a ContinuousVestingAccount.
func (cva *ContinuousVestingAccount) UnmarshalJSON(bz []byte) error {
	var alias vestingAccountPretty
	if err := json.Unmarshal(bz, &alias); err != nil {
		return err
	}

	var (
		pk  cryptotypes.PubKey
		err error
	)

	if alias.PubKey != "" {
		pk, err = legacybech32.UnmarshalPubKey(legacybech32.AccPK, alias.PubKey)
		if err != nil {
			return err
		}
	}

	cva.BaseVestingAccount = &BaseVestingAccount{
		BaseAccount:      NewBaseAccount(alias.Address, alias.Coins, pk, alias.AccountNumber, alias.Sequence),
		OriginalVesting:  alias.OriginalVesting,
		DelegatedFree:    alias.DelegatedFree,
		DelegatedVesting: alias.DelegatedVesting,
		EndTime:          alias.EndTime,
	}
	cva.StartTime = alias.StartTime

	return nil
}

func NewDelayedVestingAccountRaw(bva *BaseVestingAccount) *DelayedVestingAccount {
	return &DelayedVestingAccount{
		BaseVestingAccount: bva,
	}
}

func (dva DelayedVestingAccount) Validate() error {
	return dva.BaseVestingAccount.Validate()
}

// MarshalJSON returns the JSON representation of a DelayedVestingAccount.
func (dva DelayedVestingAccount) MarshalJSON() ([]byte, error) {
	alias := vestingAccountPretty{
		Address:          dva.Address,
		Coins:            dva.Coins,
		AccountNumber:    dva.AccountNumber,
		Sequence:         dva.Sequence,
		OriginalVesting:  dva.OriginalVesting,
		DelegatedFree:    dva.DelegatedFree,
		DelegatedVesting: dva.DelegatedVesting,
		EndTime:          dva.EndTime,
	}

	if dva.PubKey != nil {
		pks, err := legacybech32.MarshalPubKey(legacybech32.AccPK, dva.PubKey)
		if err != nil {
			return nil, err
		}

		alias.PubKey = pks
	}

	return json.Marshal(alias)
}

// UnmarshalJSON unmarshals raw JSON bytes into a DelayedVestingAccount.
func (dva *DelayedVestingAccount) UnmarshalJSON(bz []byte) error {
	var alias vestingAccountPretty
	if err := json.Unmarshal(bz, &alias); err != nil {
		return err
	}

	var (
		pk  cryptotypes.PubKey
		err error
	)

	if alias.PubKey != "" {
		pk, err = legacybech32.UnmarshalPubKey(legacybech32.AccPK, alias.PubKey)
		if err != nil {
			return err
		}
	}

	dva.BaseVestingAccount = &BaseVestingAccount{
		BaseAccount:      NewBaseAccount(alias.Address, alias.Coins, pk, alias.AccountNumber, alias.Sequence),
		OriginalVesting:  alias.OriginalVesting,
		DelegatedFree:    alias.DelegatedFree,
		DelegatedVesting: alias.DelegatedVesting,
		EndTime:          alias.EndTime,
	}

	return nil
}

func NewModuleAddress(name string) sdk.AccAddress {
	return sdk.AccAddress(tmcrypto.AddressHash([]byte(name)))
}

func NewModuleAccount(baseAccount *BaseAccount, name string, permissions ...string) *ModuleAccount {
	return &ModuleAccount{
		BaseAccount: baseAccount,
		Name:        name,
		Permissions: permissions,
	}
}

func (ma ModuleAccount) Validate() error {
	if err := ValidatePermissions(ma.Permissions...); err != nil {
		return err
	}

	if strings.TrimSpace(ma.Name) == "" {
		return errors.New("module account name cannot be blank")
	}

	if !ma.Address.Equals(sdk.AccAddress(tmcrypto.AddressHash([]byte(ma.Name)))) {
		return fmt.Errorf("address %s cannot be derived from the module name '%s'", ma.Address, ma.Name)
	}

	return ma.BaseAccount.Validate()
}

// MarshalJSON returns the JSON representation of a ModuleAccount.
func (ma ModuleAccount) MarshalJSON() ([]byte, error) {
	return json.Marshal(moduleAccountPretty{
		Address:       ma.Address,
		Coins:         ma.Coins,
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

	ma.BaseAccount = NewBaseAccount(alias.Address, alias.Coins, nil, alias.AccountNumber, alias.Sequence)
	ma.Name = alias.Name
	ma.Permissions = alias.Permissions

	return nil
}

func ValidatePermissions(permissions ...string) error {
	for _, perm := range permissions {
		if strings.TrimSpace(perm) == "" {
			return fmt.Errorf("module permission is empty")
		}
	}

	return nil
}

func SanitizeGenesisAccounts(genAccounts GenesisAccounts) GenesisAccounts {
	sort.Slice(genAccounts, func(i, j int) bool {
		return genAccounts[i].GetAccountNumber() < genAccounts[j].GetAccountNumber()
	})

	for _, acc := range genAccounts {
		if err := acc.SetCoins(acc.GetCoins().Sort()); err != nil {
			panic(err)
		}
	}

	return genAccounts
}

func ValidateGenAccounts(genAccounts GenesisAccounts) error {
	addrMap := make(map[string]bool, len(genAccounts))
	for _, acc := range genAccounts {

		// check for duplicated accounts
		addrStr := acc.GetAddress().String()
		if _, ok := addrMap[addrStr]; ok {
			return fmt.Errorf("duplicate account found in genesis state; address: %s", addrStr)
		}

		addrMap[addrStr] = true

		// check account specific validation
		if err := acc.Validate(); err != nil {
			return fmt.Errorf("invalid account found in genesis state; address: %s, error: %s", addrStr, err.Error())
		}
	}

	return nil
}

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cryptocodec.RegisterCrypto(cdc)
	cdc.RegisterInterface((*GenesisAccount)(nil), nil)
	cdc.RegisterInterface((*Account)(nil), nil)
	cdc.RegisterConcrete(&BaseAccount{}, "cosmos-sdk/Account", nil)
	cdc.RegisterConcrete(&BaseVestingAccount{}, "cosmos-sdk/BaseVestingAccount", nil)
	cdc.RegisterConcrete(&ContinuousVestingAccount{}, "cosmos-sdk/ContinuousVestingAccount", nil)
	cdc.RegisterConcrete(&DelayedVestingAccount{}, "cosmos-sdk/DelayedVestingAccount", nil)
	cdc.RegisterConcrete(&ModuleAccount{}, "cosmos-sdk/ModuleAccount", nil)
}
