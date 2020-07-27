package types

import (
	"encoding/json"
	"errors"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestexported "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"

	"github.com/tendermint/tendermint/crypto"
	"gopkg.in/yaml.v2"
)

// Compile-time type assertions
var (
	_ authexported.Account        = (*BaseVestingAccount)(nil)
	_ vestexported.VestingAccount = (*ContinuousVestingAccount)(nil)
	_ vestexported.VestingAccount = (*PeriodicVestingAccount)(nil)
	_ vestexported.VestingAccount = (*DelayedVestingAccount)(nil)
)

// Register the vesting account types on the auth module codec
func init() {
	authtypes.RegisterAccountTypeCodec(&BaseVestingAccount{}, "cosmos-sdk/BaseVestingAccount")
	authtypes.RegisterAccountTypeCodec(&ContinuousVestingAccount{}, "cosmos-sdk/ContinuousVestingAccount")
	authtypes.RegisterAccountTypeCodec(&DelayedVestingAccount{}, "cosmos-sdk/DelayedVestingAccount")
	authtypes.RegisterAccountTypeCodec(&PeriodicVestingAccount{}, "cosmos-sdk/PeriodicVestingAccount")
}

// BaseVestingAccount implements the VestingAccount interface. It contains all
// the necessary fields needed for any vesting account implementation.
type BaseVestingAccount struct {
	*authtypes.BaseAccount

	OriginalVesting  sdk.Coins `json:"original_vesting" yaml:"original_vesting"`   // coins in account upon initialization
	DelegatedFree    sdk.Coins `json:"delegated_free" yaml:"delegated_free"`       // coins that are vested and delegated
	DelegatedVesting sdk.Coins `json:"delegated_vesting" yaml:"delegated_vesting"` // coins that vesting and delegated
	EndTime          int64     `json:"end_time" yaml:"end_time"`                   // when the coins become unlocked
}

// NewBaseVestingAccount creates a new BaseVestingAccount object
func NewBaseVestingAccount(baseAccount *authtypes.BaseAccount, originalVesting sdk.Coins, endTime int64) (*BaseVestingAccount, error) {
	if (baseAccount.Coins.IsZero() && !originalVesting.IsZero()) || originalVesting.IsAnyGT(baseAccount.Coins) {
		return &BaseVestingAccount{}, errors.New("vesting amount cannot be greater than total amount")
	}
	return &BaseVestingAccount{
		BaseAccount:      baseAccount,
		OriginalVesting:  originalVesting,
		DelegatedFree:    sdk.NewCoins(),
		DelegatedVesting: sdk.NewCoins(),
		EndTime:          endTime,
	}, nil
}

// SpendableCoinsVestingAccount returns all the spendable coins for a vesting account given a
// set of vesting coins.
//
// CONTRACT: The account's coins, delegated vesting coins, vestingCoins must be
// sorted.
func (bva BaseVestingAccount) SpendableCoinsVestingAccount(vestingCoins sdk.Coins) sdk.Coins {
	var spendableCoins sdk.Coins
	bc := bva.GetCoins()

	for _, coin := range bc {
		baseAmt := coin.Amount
		vestingAmt := vestingCoins.AmountOf(coin.Denom)
		delVestingAmt := bva.DelegatedVesting.AmountOf(coin.Denom)

		// compute min((BC + DV) - V, BC) per the specification
		min := sdk.MinInt(baseAmt.Add(delVestingAmt).Sub(vestingAmt), baseAmt)
		spendableCoin := sdk.NewCoin(coin.Denom, min)

		if !spendableCoin.IsZero() {
			spendableCoins = spendableCoins.Add(spendableCoin)
		}
	}

	return spendableCoins
}

// TrackDelegation tracks a delegation amount for any given vesting account type
// given the amount of coins currently vesting.
//
// CONTRACT: The account's coins, delegation coins, vesting coins, and delegated
// vesting coins must be sorted.
func (bva *BaseVestingAccount) TrackDelegation(vestingCoins, amount sdk.Coins) {
	bc := bva.GetCoins()

	for _, coin := range amount {
		baseAmt := bc.AmountOf(coin.Denom)
		vestingAmt := vestingCoins.AmountOf(coin.Denom)
		delVestingAmt := bva.DelegatedVesting.AmountOf(coin.Denom)

		// Panic if the delegation amount is zero or if the base coins does not
		// exceed the desired delegation amount.
		if coin.Amount.IsZero() || baseAmt.LT(coin.Amount) {
			panic("delegation attempt with zero coins or insufficient funds")
		}

		// compute x and y per the specification, where:
		// X := min(max(V - DV, 0), D)
		// Y := D - X
		x := sdk.MinInt(sdk.MaxInt(vestingAmt.Sub(delVestingAmt), sdk.ZeroInt()), coin.Amount)
		y := coin.Amount.Sub(x)

		if !x.IsZero() {
			xCoin := sdk.NewCoin(coin.Denom, x)
			bva.DelegatedVesting = bva.DelegatedVesting.Add(xCoin)
		}

		if !y.IsZero() {
			yCoin := sdk.NewCoin(coin.Denom, y)
			bva.DelegatedFree = bva.DelegatedFree.Add(yCoin)
		}
	}
}

// TrackUndelegation tracks an undelegation amount by setting the necessary
// values by which delegated vesting and delegated vesting need to decrease and
// by which amount the base coins need to increase.
//
// NOTE: The undelegation (bond refund) amount may exceed the delegated
// vesting (bond) amount due to the way undelegation truncates the bond refund,
// which can increase the validator's exchange rate (tokens/shares) slightly if
// the undelegated tokens are non-integral.
//
// CONTRACT: The account's coins and undelegation coins must be sorted.
func (bva *BaseVestingAccount) TrackUndelegation(amount sdk.Coins) {
	for _, coin := range amount {
		// panic if the undelegation amount is zero
		if coin.Amount.IsZero() {
			panic("undelegation attempt with zero coins")
		}
		delegatedFree := bva.DelegatedFree.AmountOf(coin.Denom)
		delegatedVesting := bva.DelegatedVesting.AmountOf(coin.Denom)

		// compute x and y per the specification, where:
		// X := min(DF, D)
		// Y := min(DV, D - X)
		x := sdk.MinInt(delegatedFree, coin.Amount)
		y := sdk.MinInt(delegatedVesting, coin.Amount.Sub(x))

		if !x.IsZero() {
			xCoin := sdk.NewCoin(coin.Denom, x)
			bva.DelegatedFree = bva.DelegatedFree.Sub(sdk.Coins{xCoin})
		}

		if !y.IsZero() {
			yCoin := sdk.NewCoin(coin.Denom, y)
			bva.DelegatedVesting = bva.DelegatedVesting.Sub(sdk.Coins{yCoin})
		}
	}
}

// GetOriginalVesting returns a vesting account's original vesting amount
func (bva BaseVestingAccount) GetOriginalVesting() sdk.Coins {
	return bva.OriginalVesting
}

// GetDelegatedFree returns a vesting account's delegation amount that is not
// vesting.
func (bva BaseVestingAccount) GetDelegatedFree() sdk.Coins {
	return bva.DelegatedFree
}

// GetDelegatedVesting returns a vesting account's delegation amount that is
// still vesting.
func (bva BaseVestingAccount) GetDelegatedVesting() sdk.Coins {
	return bva.DelegatedVesting
}

// GetEndTime returns a vesting account's end time
func (bva BaseVestingAccount) GetEndTime() int64 {
	return bva.EndTime
}

// Validate checks for errors on the account fields
func (bva BaseVestingAccount) Validate() error {
	if !(bva.DelegatedVesting.IsAllLTE(bva.OriginalVesting)) {
		return errors.New("delegated vesting amount cannot be greater than original vesting amount")
	}
	return bva.BaseAccount.Validate()
}

type vestingAccountPretty struct {
	Address          sdk.AccAddress `json:"address" yaml:"address"`
	Coins            sdk.Coins      `json:"coins" yaml:"coins"`
	PubKey           string         `json:"public_key" yaml:"public_key"`
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

func (bva BaseVestingAccount) String() string {
	out, _ := bva.MarshalYAML()
	return out.(string)
}

// MarshalYAML returns the YAML representation of a BaseVestingAccount.
func (bva BaseVestingAccount) MarshalYAML() (interface{}, error) {
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
		pks, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, bva.PubKey)
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
		pks, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, bva.PubKey)
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
		pk  crypto.PubKey
		err error
	)

	if alias.PubKey != "" {
		pk, err = sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeAccPub, alias.PubKey)
		if err != nil {
			return err
		}
	}

	bva.BaseAccount = authtypes.NewBaseAccount(alias.Address, alias.Coins, pk, alias.AccountNumber, alias.Sequence)
	bva.OriginalVesting = alias.OriginalVesting
	bva.DelegatedFree = alias.DelegatedFree
	bva.DelegatedVesting = alias.DelegatedVesting
	bva.EndTime = alias.EndTime

	return nil
}

//-----------------------------------------------------------------------------
// Continuous Vesting Account

var _ vestexported.VestingAccount = (*ContinuousVestingAccount)(nil)
var _ authexported.GenesisAccount = (*ContinuousVestingAccount)(nil)

// ContinuousVestingAccount implements the VestingAccount interface. It
// continuously vests by unlocking coins linearly with respect to time.
type ContinuousVestingAccount struct {
	*BaseVestingAccount

	StartTime int64 `json:"start_time" yaml:"start_time"` // when the coins start to vest
}

// NewContinuousVestingAccountRaw creates a new ContinuousVestingAccount object from BaseVestingAccount
func NewContinuousVestingAccountRaw(bva *BaseVestingAccount, startTime int64) *ContinuousVestingAccount {
	return &ContinuousVestingAccount{
		BaseVestingAccount: bva,
		StartTime:          startTime,
	}
}

// NewContinuousVestingAccount returns a new ContinuousVestingAccount
func NewContinuousVestingAccount(baseAcc *authtypes.BaseAccount, startTime, endTime int64) *ContinuousVestingAccount {
	baseVestingAcc := &BaseVestingAccount{
		BaseAccount:     baseAcc,
		OriginalVesting: baseAcc.Coins,
		EndTime:         endTime,
	}

	return &ContinuousVestingAccount{
		StartTime:          startTime,
		BaseVestingAccount: baseVestingAcc,
	}
}

// GetVestedCoins returns the total number of vested coins. If no coins are vested,
// nil is returned.
func (cva ContinuousVestingAccount) GetVestedCoins(blockTime time.Time) sdk.Coins {
	var vestedCoins sdk.Coins

	// We must handle the case where the start time for a vesting account has
	// been set into the future or when the start of the chain is not exactly
	// known.
	if blockTime.Unix() <= cva.StartTime {
		return vestedCoins
	} else if blockTime.Unix() >= cva.EndTime {
		return cva.OriginalVesting
	}

	// calculate the vesting scalar
	x := blockTime.Unix() - cva.StartTime
	y := cva.EndTime - cva.StartTime
	s := sdk.NewDec(x).Quo(sdk.NewDec(y))

	for _, ovc := range cva.OriginalVesting {
		vestedAmt := ovc.Amount.ToDec().Mul(s).RoundInt()
		vestedCoins = append(vestedCoins, sdk.NewCoin(ovc.Denom, vestedAmt))
	}

	return vestedCoins
}

// GetVestingCoins returns the total number of vesting coins. If no coins are
// vesting, nil is returned.
func (cva ContinuousVestingAccount) GetVestingCoins(blockTime time.Time) sdk.Coins {
	return cva.OriginalVesting.Sub(cva.GetVestedCoins(blockTime))
}

// SpendableCoins returns the total number of spendable coins per denom for a
// continuous vesting account.
func (cva ContinuousVestingAccount) SpendableCoins(blockTime time.Time) sdk.Coins {
	return cva.BaseVestingAccount.SpendableCoinsVestingAccount(cva.GetVestingCoins(blockTime))
}

// TrackDelegation tracks a desired delegation amount by setting the appropriate
// values for the amount of delegated vesting, delegated free, and reducing the
// overall amount of base coins.
func (cva *ContinuousVestingAccount) TrackDelegation(blockTime time.Time, amount sdk.Coins) {
	cva.BaseVestingAccount.TrackDelegation(cva.GetVestingCoins(blockTime), amount)
}

// GetStartTime returns the time when vesting starts for a continuous vesting
// account.
func (cva ContinuousVestingAccount) GetStartTime() int64 {
	return cva.StartTime
}

// Validate checks for errors on the account fields
func (cva ContinuousVestingAccount) Validate() error {
	if cva.GetStartTime() >= cva.GetEndTime() {
		return errors.New("vesting start-time cannot be before end-time")
	}

	return cva.BaseVestingAccount.Validate()
}

func (cva ContinuousVestingAccount) String() string {
	out, _ := cva.MarshalYAML()
	return out.(string)
}

// MarshalYAML returns the YAML representation of a ContinuousVestingAccount.
func (cva ContinuousVestingAccount) MarshalYAML() (interface{}, error) {
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
		pks, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, cva.PubKey)
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
		pks, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, cva.PubKey)
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
		pk  crypto.PubKey
		err error
	)

	if alias.PubKey != "" {
		pk, err = sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeAccPub, alias.PubKey)
		if err != nil {
			return err
		}
	}

	cva.BaseVestingAccount = &BaseVestingAccount{
		BaseAccount:      authtypes.NewBaseAccount(alias.Address, alias.Coins, pk, alias.AccountNumber, alias.Sequence),
		OriginalVesting:  alias.OriginalVesting,
		DelegatedFree:    alias.DelegatedFree,
		DelegatedVesting: alias.DelegatedVesting,
		EndTime:          alias.EndTime,
	}
	cva.StartTime = alias.StartTime

	return nil
}

//-----------------------------------------------------------------------------
// Periodic Vesting Account

var _ vestexported.VestingAccount = (*PeriodicVestingAccount)(nil)
var _ authexported.GenesisAccount = (*PeriodicVestingAccount)(nil)

// PeriodicVestingAccount implements the VestingAccount interface. It
// periodically vests by unlocking coins during each specified period
type PeriodicVestingAccount struct {
	*BaseVestingAccount
	StartTime      int64   `json:"start_time" yaml:"start_time"`           // when the coins start to vest
	VestingPeriods Periods `json:"vesting_periods" yaml:"vesting_periods"` // the vesting schedule
}

// NewPeriodicVestingAccountRaw creates a new PeriodicVestingAccount object from BaseVestingAccount
func NewPeriodicVestingAccountRaw(bva *BaseVestingAccount, startTime int64, periods Periods) *PeriodicVestingAccount {
	return &PeriodicVestingAccount{
		BaseVestingAccount: bva,
		StartTime:          startTime,
		VestingPeriods:     periods,
	}
}

// NewPeriodicVestingAccount returns a new PeriodicVestingAccount
func NewPeriodicVestingAccount(baseAcc *authtypes.BaseAccount, startTime int64, periods Periods) *PeriodicVestingAccount {
	endTime := startTime
	for _, p := range periods {
		endTime += p.Length
	}
	baseVestingAcc := &BaseVestingAccount{
		BaseAccount:     baseAcc,
		OriginalVesting: baseAcc.Coins,
		EndTime:         endTime,
	}

	return &PeriodicVestingAccount{
		BaseVestingAccount: baseVestingAcc,
		StartTime:          startTime,
		VestingPeriods:     periods,
	}
}

// GetVestedCoins returns the total number of vested coins. If no coins are vested,
// nil is returned.
func (pva PeriodicVestingAccount) GetVestedCoins(blockTime time.Time) sdk.Coins {
	var vestedCoins sdk.Coins

	// We must handle the case where the start time for a vesting account has
	// been set into the future or when the start of the chain is not exactly
	// known.
	if blockTime.Unix() <= pva.StartTime {
		return vestedCoins
	} else if blockTime.Unix() >= pva.EndTime {
		return pva.OriginalVesting
	}

	// track the start time of the next period
	currentPeriodStartTime := pva.StartTime
	// for each period, if the period is over, add those coins as vested and check the next period.
	for _, period := range pva.VestingPeriods {
		x := blockTime.Unix() - currentPeriodStartTime
		if x < period.Length {
			break
		}
		vestedCoins = vestedCoins.Add(period.Amount...)
		// Update the start time of the next period
		currentPeriodStartTime += period.Length
	}
	return vestedCoins
}

// GetVestingCoins returns the total number of vesting coins. If no coins are
// vesting, nil is returned.
func (pva PeriodicVestingAccount) GetVestingCoins(blockTime time.Time) sdk.Coins {
	return pva.OriginalVesting.Sub(pva.GetVestedCoins(blockTime))
}

// SpendableCoins returns the total number of spendable coins per denom for a
// periodic vesting account.
func (pva PeriodicVestingAccount) SpendableCoins(blockTime time.Time) sdk.Coins {
	return pva.BaseVestingAccount.SpendableCoinsVestingAccount(pva.GetVestingCoins(blockTime))
}

// TrackDelegation tracks a desired delegation amount by setting the appropriate
// values for the amount of delegated vesting, delegated free, and reducing the
// overall amount of base coins.
func (pva *PeriodicVestingAccount) TrackDelegation(blockTime time.Time, amount sdk.Coins) {
	pva.BaseVestingAccount.TrackDelegation(pva.GetVestingCoins(blockTime), amount)
}

// GetStartTime returns the time when vesting starts for a periodic vesting
// account.
func (pva PeriodicVestingAccount) GetStartTime() int64 {
	return pva.StartTime
}

// GetVestingPeriods returns vesting periods associated with periodic vesting account.
func (pva PeriodicVestingAccount) GetVestingPeriods() Periods {
	return pva.VestingPeriods
}

// Validate checks for errors on the account fields
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

func (pva PeriodicVestingAccount) String() string {
	out, _ := pva.MarshalYAML()
	return out.(string)
}

// MarshalYAML returns the YAML representation of a PeriodicVestingAccount.
func (pva PeriodicVestingAccount) MarshalYAML() (interface{}, error) {
	alias := vestingAccountPretty{
		Address:          pva.Address,
		Coins:            pva.Coins,
		AccountNumber:    pva.AccountNumber,
		Sequence:         pva.Sequence,
		OriginalVesting:  pva.OriginalVesting,
		DelegatedFree:    pva.DelegatedFree,
		DelegatedVesting: pva.DelegatedVesting,
		EndTime:          pva.EndTime,
		StartTime:        pva.StartTime,
		VestingPeriods:   pva.VestingPeriods,
	}

	if pva.PubKey != nil {
		pks, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, pva.PubKey)
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

// MarshalJSON returns the JSON representation of a PeriodicVestingAccount.
func (pva PeriodicVestingAccount) MarshalJSON() ([]byte, error) {
	alias := vestingAccountPretty{
		Address:          pva.Address,
		Coins:            pva.Coins,
		AccountNumber:    pva.AccountNumber,
		Sequence:         pva.Sequence,
		OriginalVesting:  pva.OriginalVesting,
		DelegatedFree:    pva.DelegatedFree,
		DelegatedVesting: pva.DelegatedVesting,
		EndTime:          pva.EndTime,
		StartTime:        pva.StartTime,
		VestingPeriods:   pva.VestingPeriods,
	}

	if pva.PubKey != nil {
		pks, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, pva.PubKey)
		if err != nil {
			return nil, err
		}

		alias.PubKey = pks
	}

	return json.Marshal(alias)
}

// UnmarshalJSON unmarshals raw JSON bytes into a PeriodicVestingAccount.
func (pva *PeriodicVestingAccount) UnmarshalJSON(bz []byte) error {
	var alias vestingAccountPretty
	if err := json.Unmarshal(bz, &alias); err != nil {
		return err
	}

	var (
		pk  crypto.PubKey
		err error
	)

	if alias.PubKey != "" {
		pk, err = sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeAccPub, alias.PubKey)
		if err != nil {
			return err
		}
	}

	pva.BaseVestingAccount = &BaseVestingAccount{
		BaseAccount:      authtypes.NewBaseAccount(alias.Address, alias.Coins, pk, alias.AccountNumber, alias.Sequence),
		OriginalVesting:  alias.OriginalVesting,
		DelegatedFree:    alias.DelegatedFree,
		DelegatedVesting: alias.DelegatedVesting,
		EndTime:          alias.EndTime,
	}
	pva.StartTime = alias.StartTime
	pva.VestingPeriods = alias.VestingPeriods

	return nil
}

//-----------------------------------------------------------------------------
// Delayed Vesting Account

var _ vestexported.VestingAccount = (*DelayedVestingAccount)(nil)
var _ authexported.GenesisAccount = (*DelayedVestingAccount)(nil)

// DelayedVestingAccount implements the VestingAccount interface. It vests all
// coins after a specific time, but non prior. In other words, it keeps them
// locked until a specified time.
type DelayedVestingAccount struct {
	*BaseVestingAccount
}

// NewDelayedVestingAccountRaw creates a new DelayedVestingAccount object from BaseVestingAccount
func NewDelayedVestingAccountRaw(bva *BaseVestingAccount) *DelayedVestingAccount {
	return &DelayedVestingAccount{
		BaseVestingAccount: bva,
	}
}

// NewDelayedVestingAccount returns a DelayedVestingAccount
func NewDelayedVestingAccount(baseAcc *authtypes.BaseAccount, endTime int64) *DelayedVestingAccount {
	baseVestingAcc := &BaseVestingAccount{
		BaseAccount:     baseAcc,
		OriginalVesting: baseAcc.Coins,
		EndTime:         endTime,
	}

	return &DelayedVestingAccount{baseVestingAcc}
}

// GetVestedCoins returns the total amount of vested coins for a delayed vesting
// account. All coins are only vested once the schedule has elapsed.
func (dva DelayedVestingAccount) GetVestedCoins(blockTime time.Time) sdk.Coins {
	if blockTime.Unix() >= dva.EndTime {
		return dva.OriginalVesting
	}

	return nil
}

// GetVestingCoins returns the total number of vesting coins for a delayed
// vesting account.
func (dva DelayedVestingAccount) GetVestingCoins(blockTime time.Time) sdk.Coins {
	return dva.OriginalVesting.Sub(dva.GetVestedCoins(blockTime))
}

// SpendableCoins returns the total number of spendable coins for a delayed
// vesting account.
func (dva DelayedVestingAccount) SpendableCoins(blockTime time.Time) sdk.Coins {
	return dva.BaseVestingAccount.SpendableCoinsVestingAccount(dva.GetVestingCoins(blockTime))
}

// TrackDelegation tracks a desired delegation amount by setting the appropriate
// values for the amount of delegated vesting, delegated free, and reducing the
// overall amount of base coins.
func (dva *DelayedVestingAccount) TrackDelegation(blockTime time.Time, amount sdk.Coins) {
	dva.BaseVestingAccount.TrackDelegation(dva.GetVestingCoins(blockTime), amount)
}

// GetStartTime returns zero since a delayed vesting account has no start time.
func (dva DelayedVestingAccount) GetStartTime() int64 {
	return 0
}

// Validate checks for errors on the account fields
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
		pks, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, dva.PubKey)
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
		pk  crypto.PubKey
		err error
	)

	if alias.PubKey != "" {
		pk, err = sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeAccPub, alias.PubKey)
		if err != nil {
			return err
		}
	}

	dva.BaseVestingAccount = &BaseVestingAccount{
		BaseAccount:      authtypes.NewBaseAccount(alias.Address, alias.Coins, pk, alias.AccountNumber, alias.Sequence),
		OriginalVesting:  alias.OriginalVesting,
		DelegatedFree:    alias.DelegatedFree,
		DelegatedVesting: alias.DelegatedVesting,
		EndTime:          alias.EndTime,
	}

	return nil
}
