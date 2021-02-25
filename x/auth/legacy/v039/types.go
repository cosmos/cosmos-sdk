package v039

// DONTCOVER

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	tmcrypto "github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v034auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v034"
	v038auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v038"
)

const (
	ModuleName = "auth"
)

type (
	GenesisState struct {
		Params   v034auth.Params          `json:"params" yaml:"params"`
		Accounts v038auth.GenesisAccounts `json:"accounts" yaml:"accounts"`
	}

	BaseAccount struct {
		Address       sdk.AccAddress     `json:"address" yaml:"address"`
		Coins         sdk.Coins          `json:"coins,omitempty" yaml:"coins,omitempty"`
		PubKey        cryptotypes.PubKey `json:"public_key" yaml:"public_key"`
		AccountNumber uint64             `json:"account_number" yaml:"account_number"`
		Sequence      uint64             `json:"sequence" yaml:"sequence"`
	}

	BaseVestingAccount struct {
		*BaseAccount

		OriginalVesting  sdk.Coins `json:"original_vesting"`
		DelegatedFree    sdk.Coins `json:"delegated_free"`
		DelegatedVesting sdk.Coins `json:"delegated_vesting"`

		EndTime int64 `json:"end_time"`
	}

	vestingAccountJSON struct {
		Address          sdk.AccAddress     `json:"address" yaml:"address"`
		Coins            sdk.Coins          `json:"coins,omitempty" yaml:"coins"`
		PubKey           cryptotypes.PubKey `json:"public_key" yaml:"public_key"`
		AccountNumber    uint64             `json:"account_number" yaml:"account_number"`
		Sequence         uint64             `json:"sequence" yaml:"sequence"`
		OriginalVesting  sdk.Coins          `json:"original_vesting" yaml:"original_vesting"`
		DelegatedFree    sdk.Coins          `json:"delegated_free" yaml:"delegated_free"`
		DelegatedVesting sdk.Coins          `json:"delegated_vesting" yaml:"delegated_vesting"`
		EndTime          int64              `json:"end_time" yaml:"end_time"`

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

func NewGenesisState(params v034auth.Params, accounts v038auth.GenesisAccounts) GenesisState {
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
		Coins:            bva.Coins,
		PubKey:           bva.PubKey,
		AccountNumber:    bva.AccountNumber,
		Sequence:         bva.Sequence,
		OriginalVesting:  bva.OriginalVesting,
		DelegatedFree:    bva.DelegatedFree,
		DelegatedVesting: bva.DelegatedVesting,
		EndTime:          bva.EndTime,
	}

	return legacy.Cdc.MarshalJSON(alias)
}

func (bva *BaseVestingAccount) UnmarshalJSON(bz []byte) error {
	var alias vestingAccountJSON
	if err := legacy.Cdc.UnmarshalJSON(bz, &alias); err != nil {
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
		Coins:            cva.Coins,
		PubKey:           cva.PubKey,
		AccountNumber:    cva.AccountNumber,
		Sequence:         cva.Sequence,
		OriginalVesting:  cva.OriginalVesting,
		DelegatedFree:    cva.DelegatedFree,
		DelegatedVesting: cva.DelegatedVesting,
		EndTime:          cva.EndTime,
		StartTime:        cva.StartTime,
	}

	return legacy.Cdc.MarshalJSON(alias)
}

func (cva *ContinuousVestingAccount) UnmarshalJSON(bz []byte) error {
	var alias vestingAccountJSON
	if err := legacy.Cdc.UnmarshalJSON(bz, &alias); err != nil {
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

func NewDelayedVestingAccountRaw(bva *BaseVestingAccount) *DelayedVestingAccount {
	return &DelayedVestingAccount{
		BaseVestingAccount: bva,
	}
}

func (dva DelayedVestingAccount) Validate() error {
	return dva.BaseVestingAccount.Validate()
}

func (dva DelayedVestingAccount) MarshalJSON() ([]byte, error) {
	alias := vestingAccountJSON{
		Address:          dva.Address,
		Coins:            dva.Coins,
		PubKey:           dva.PubKey,
		AccountNumber:    dva.AccountNumber,
		Sequence:         dva.Sequence,
		OriginalVesting:  dva.OriginalVesting,
		DelegatedFree:    dva.DelegatedFree,
		DelegatedVesting: dva.DelegatedVesting,
		EndTime:          dva.EndTime,
	}

	return legacy.Cdc.MarshalJSON(alias)
}

// UnmarshalJSON unmarshals raw JSON bytes into a DelayedVestingAccount.
func (dva *DelayedVestingAccount) UnmarshalJSON(bz []byte) error {
	var alias vestingAccountJSON
	if err := legacy.Cdc.UnmarshalJSON(bz, &alias); err != nil {
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
		Coins:            pva.Coins,
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

	return legacy.Cdc.MarshalJSON(alias)
}

// UnmarshalJSON unmarshals raw JSON bytes into a PeriodicVestingAccount.
func (pva *PeriodicVestingAccount) UnmarshalJSON(bz []byte) error {
	var alias vestingAccountJSON
	if err := legacy.Cdc.UnmarshalJSON(bz, &alias); err != nil {
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

func NewModuleAccount(baseAccount *BaseAccount, name string, permissions ...string) *ModuleAccount {
	return &ModuleAccount{
		BaseAccount: baseAccount,
		Name:        name,
		Permissions: permissions,
	}
}

func (ma ModuleAccount) Validate() error {
	if err := v038auth.ValidatePermissions(ma.Permissions...); err != nil {
		return err
	}

	if strings.TrimSpace(ma.Name) == "" {
		return errors.New("module account name cannot be blank")
	}

	if x := sdk.AccAddress(tmcrypto.AddressHash([]byte(ma.Name))); !ma.Address.Equals(x) {
		return fmt.Errorf("address %s cannot be derived from the module name '%s'; expected: %s", ma.Address, ma.Name, x)
	}

	return ma.BaseAccount.Validate()
}

// MarshalJSON returns the JSON representation of a ModuleAccount.
func (ma ModuleAccount) MarshalJSON() ([]byte, error) {
	return legacy.Cdc.MarshalJSON(moduleAccountPretty{
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
	if err := legacy.Cdc.UnmarshalJSON(bz, &alias); err != nil {
		return err
	}

	ma.BaseAccount = NewBaseAccount(alias.Address, alias.Coins, nil, alias.AccountNumber, alias.Sequence)
	ma.Name = alias.Name
	ma.Permissions = alias.Permissions

	return nil
}

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cryptocodec.RegisterCrypto(cdc)
	cdc.RegisterInterface((*v038auth.GenesisAccount)(nil), nil)
	cdc.RegisterInterface((*v038auth.Account)(nil), nil)
	cdc.RegisterConcrete(&BaseAccount{}, "cosmos-sdk/Account", nil)
	cdc.RegisterConcrete(&BaseVestingAccount{}, "cosmos-sdk/BaseVestingAccount", nil)
	cdc.RegisterConcrete(&ContinuousVestingAccount{}, "cosmos-sdk/ContinuousVestingAccount", nil)
	cdc.RegisterConcrete(&DelayedVestingAccount{}, "cosmos-sdk/DelayedVestingAccount", nil)
	cdc.RegisterConcrete(&PeriodicVestingAccount{}, "cosmos-sdk/PeriodicVestingAccount", nil)
	cdc.RegisterConcrete(&ModuleAccount{}, "cosmos-sdk/ModuleAccount", nil)
}
