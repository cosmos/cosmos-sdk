package v039

// DONTCOVER
// nolint

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v034auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_34"
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
		Address       sdk.AccAddress `json:"address" yaml:"address"`
		Coins         sdk.Coins      `json:"coins,omitempty" yaml:"coins,omitempty"`
		PubKey        crypto.PubKey  `json:"public_key" yaml:"public_key"`
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

	vestingAccountJSON struct {
		Address          sdk.AccAddress `json:"address" yaml:"address"`
		Coins            sdk.Coins      `json:"coins,omitempty" yaml:"coins"`
		PubKey           crypto.PubKey  `json:"public_key" yaml:"public_key"`
		AccountNumber    uint64         `json:"account_number" yaml:"account_number"`
		Sequence         uint64         `json:"sequence" yaml:"sequence"`
		OriginalVesting  sdk.Coins      `json:"original_vesting" yaml:"original_vesting"`
		DelegatedFree    sdk.Coins      `json:"delegated_free" yaml:"delegated_free"`
		DelegatedVesting sdk.Coins      `json:"delegated_vesting" yaml:"delegated_vesting"`
		EndTime          int64          `json:"end_time" yaml:"end_time"`

		// custom fields based on concrete vesting type which can be omitted
		StartTime      int64   `json:"start_time,omitempty" yaml:"start_time,omitempty"`
		VestingPeriods Periods `json:"vesting_periods,omitempty" yaml:"vesting_periods,omitempty"`
	}

	ContinuousVestingAccount struct {
		*BaseVestingAccount

		StartTime int64 `json:"start_time"`
	}

	DelayedVestingAccount struct {
		*BaseVestingAccount
	}

	Period struct {
		Length int64     `json:"length" yaml:"length"` // length of the period, in seconds
		Amount sdk.Coins `json:"amount" yaml:"amount"` // amount of coins vesting during this period
	}

	Periods []Period

	PeriodicVestingAccount struct {
		*BaseVestingAccount
		StartTime      int64   `json:"start_time" yaml:"start_time"`           // when the coins start to vest
		VestingPeriods Periods `json:"vesting_periods" yaml:"vesting_periods"` // the vesting schedule
	}

	ModuleAccount struct {
		*BaseAccount

		Name        string   `json:"name" yaml:"name"`
		Permissions []string `json:"permissions" yaml:"permissions"`
	}

	moduleAccountPretty struct {
		Address       sdk.AccAddress `json:"address" yaml:"address"`
		Coins         sdk.Coins      `json:"coins,omitempty" yaml:"coins"`
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
	address sdk.AccAddress, coins sdk.Coins, pk crypto.PubKey, accountNumber, sequence uint64,
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

func (bva BaseVestingAccount) MarshalJSON() ([]byte, error) {
	alias := vestingAccountJSON{
		Address:          bva.Address,
		PubKey:           bva.PubKey,
		AccountNumber:    bva.AccountNumber,
		Sequence:         bva.Sequence,
		OriginalVesting:  bva.OriginalVesting,
		DelegatedFree:    bva.DelegatedFree,
		DelegatedVesting: bva.DelegatedVesting,
		EndTime:          bva.EndTime,
	}

	return codec.Cdc.MarshalJSON(alias)
}

func (bva *BaseVestingAccount) UnmarshalJSON(bz []byte) error {
	var alias vestingAccountJSON
	if err := codec.Cdc.UnmarshalJSON(bz, &alias); err != nil {
		return err
	}

	bva.BaseAccount = NewBaseAccount(alias.Address, alias.Coins, alias.PubKey, alias.AccountNumber, alias.Sequence)
	bva.OriginalVesting = alias.OriginalVesting
	bva.DelegatedFree = alias.DelegatedFree
	bva.DelegatedVesting = alias.DelegatedVesting
	bva.EndTime = alias.EndTime

	return nil
}

func (bva BaseVestingAccount) GetEndTime() int64 {
	return bva.EndTime
}

func (bva BaseVestingAccount) Validate() error {
	if (bva.Coins.IsZero() && !bva.OriginalVesting.IsZero()) ||
		bva.OriginalVesting.IsAnyGT(bva.Coins) {
		return errors.New("vesting amount cannot be greater than total amount")
	}

	return bva.BaseAccount.Validate()
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

func (cva ContinuousVestingAccount) MarshalJSON() ([]byte, error) {
	alias := vestingAccountJSON{
		Address:          cva.Address,
		PubKey:           cva.PubKey,
		AccountNumber:    cva.AccountNumber,
		Sequence:         cva.Sequence,
		OriginalVesting:  cva.OriginalVesting,
		DelegatedFree:    cva.DelegatedFree,
		DelegatedVesting: cva.DelegatedVesting,
		EndTime:          cva.EndTime,
		StartTime:        cva.StartTime,
	}

	return codec.Cdc.MarshalJSON(alias)
}

func (cva *ContinuousVestingAccount) UnmarshalJSON(bz []byte) error {
	var alias vestingAccountJSON
	if err := codec.Cdc.UnmarshalJSON(bz, &alias); err != nil {
		return err
	}

	cva.BaseVestingAccount = &BaseVestingAccount{
		BaseAccount:      NewBaseAccount(alias.Address, alias.Coins, alias.PubKey, alias.AccountNumber, alias.Sequence),
		OriginalVesting:  alias.OriginalVesting,
		DelegatedFree:    alias.DelegatedFree,
		DelegatedVesting: alias.DelegatedVesting,
		EndTime:          alias.EndTime,
	}
	cva.StartTime = alias.StartTime

	return nil
}

func (dva DelayedVestingAccount) Validate() error {
	return dva.BaseVestingAccount.Validate()
}

func (dva DelayedVestingAccount) MarshalJSON() ([]byte, error) {
	alias := vestingAccountJSON{
		Address:          dva.Address,
		PubKey:           dva.PubKey,
		AccountNumber:    dva.AccountNumber,
		Sequence:         dva.Sequence,
		OriginalVesting:  dva.OriginalVesting,
		DelegatedFree:    dva.DelegatedFree,
		DelegatedVesting: dva.DelegatedVesting,
		EndTime:          dva.EndTime,
	}

	return codec.Cdc.MarshalJSON(alias)
}

// UnmarshalJSON unmarshals raw JSON bytes into a DelayedVestingAccount.
func (dva *DelayedVestingAccount) UnmarshalJSON(bz []byte) error {
	var alias vestingAccountJSON
	if err := codec.Cdc.UnmarshalJSON(bz, &alias); err != nil {
		return err
	}

	dva.BaseVestingAccount = &BaseVestingAccount{
		BaseAccount:      NewBaseAccount(alias.Address, alias.Coins, alias.PubKey, alias.AccountNumber, alias.Sequence),
		OriginalVesting:  alias.OriginalVesting,
		DelegatedFree:    alias.DelegatedFree,
		DelegatedVesting: alias.DelegatedVesting,
		EndTime:          alias.EndTime,
	}

	return nil
}

func (pva PeriodicVestingAccount) GetStartTime() int64 {
	return pva.StartTime
}

func (pva PeriodicVestingAccount) Validate() error {
	if pva.GetStartTime() >= pva.GetEndTime() {
		return errors.New("vesting start-time cannot be before end-time")
	}
	endTime := pva.StartTime
	originalVesting := sdk.NewCoins()
	for _, p := range pva.VestingPeriods {
		endTime += p.Length
		originalVesting = originalVesting.Add(p.Amount...)
	}
	if endTime != pva.EndTime {
		return errors.New("vesting end time does not match length of all vesting periods")
	}
	if !originalVesting.IsEqual(pva.OriginalVesting) {
		return errors.New("original vesting coins does not match the sum of all coins in vesting periods")
	}

	return pva.BaseVestingAccount.Validate()
}

func (pva PeriodicVestingAccount) MarshalJSON() ([]byte, error) {
	alias := vestingAccountJSON{
		Address:          pva.Address,
		PubKey:           pva.PubKey,
		AccountNumber:    pva.AccountNumber,
		Sequence:         pva.Sequence,
		OriginalVesting:  pva.OriginalVesting,
		DelegatedFree:    pva.DelegatedFree,
		DelegatedVesting: pva.DelegatedVesting,
		EndTime:          pva.EndTime,
		StartTime:        pva.StartTime,
		VestingPeriods:   pva.VestingPeriods,
	}

	return codec.Cdc.MarshalJSON(alias)
}

// UnmarshalJSON unmarshals raw JSON bytes into a PeriodicVestingAccount.
func (pva *PeriodicVestingAccount) UnmarshalJSON(bz []byte) error {
	var alias vestingAccountJSON
	if err := codec.Cdc.UnmarshalJSON(bz, &alias); err != nil {
		return err
	}

	pva.BaseVestingAccount = &BaseVestingAccount{
		BaseAccount:      NewBaseAccount(alias.Address, alias.Coins, alias.PubKey, alias.AccountNumber, alias.Sequence),
		OriginalVesting:  alias.OriginalVesting,
		DelegatedFree:    alias.DelegatedFree,
		DelegatedVesting: alias.DelegatedVesting,
		EndTime:          alias.EndTime,
	}
	pva.StartTime = alias.StartTime
	pva.VestingPeriods = alias.VestingPeriods

	return nil
}

func (ma ModuleAccount) Validate() error {
	if err := validatePermissions(ma.Permissions...); err != nil {
		return err
	}

	if strings.TrimSpace(ma.Name) == "" {
		return errors.New("module account name cannot be blank")
	}

	if !ma.Address.Equals(sdk.AccAddress(crypto.AddressHash([]byte(ma.Name)))) {
		return fmt.Errorf("address %s cannot be derived from the module name '%s'", ma.Address, ma.Name)
	}

	return ma.BaseAccount.Validate()
}

// MarshalJSON returns the JSON representation of a ModuleAccount.
func (ma ModuleAccount) MarshalJSON() ([]byte, error) {
	return codec.Cdc.MarshalJSON(moduleAccountPretty{
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
	if err := codec.Cdc.UnmarshalJSON(bz, &alias); err != nil {
		return err
	}

	ma.BaseAccount = NewBaseAccount(alias.Address, alias.Coins, nil, alias.AccountNumber, alias.Sequence)
	ma.Name = alias.Name
	ma.Permissions = alias.Permissions

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

func validatePermissions(permissions ...string) error {
	for _, perm := range permissions {
		if strings.TrimSpace(perm) == "" {
			return fmt.Errorf("module permission is empty")
		}
	}

	return nil
}

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*GenesisAccount)(nil), nil)
	cdc.RegisterInterface((*Account)(nil), nil)
	cdc.RegisterConcrete(&BaseAccount{}, "cosmos-sdk/BaseAccount", nil)
	cdc.RegisterConcrete(&BaseVestingAccount{}, "cosmos-sdk/BaseVestingAccount", nil)
	cdc.RegisterConcrete(&ContinuousVestingAccount{}, "cosmos-sdk/ContinuousVestingAccount", nil)
	cdc.RegisterConcrete(&DelayedVestingAccount{}, "cosmos-sdk/DelayedVestingAccount", nil)
	cdc.RegisterConcrete(&PeriodicVestingAccount{}, "cosmos-sdk/PeriodicVestingAccount", nil)
	cdc.RegisterConcrete(&ModuleAccount{}, "cosmos-sdk/ModuleAccount", nil)
}
