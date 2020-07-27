package v039

// DONTCOVER
// nolint

import (
	"bytes"
	"errors"
	"fmt"
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
)

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

func (cva ContinuousVestingAccount) Validate() error {
	if cva.StartTime >= cva.EndTime {
		return errors.New("vesting start-time cannot be before end-time")
	}

	return cva.BaseVestingAccount.Validate()
}

func (dva DelayedVestingAccount) Validate() error {
	return dva.BaseVestingAccount.Validate()
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
