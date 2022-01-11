package types

import (
	"errors"
	"math"
	"time"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestexported "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
)

// Compile-time type assertions
var (
	_ sdk.AccountI                = (*BaseVestingAccount)(nil)
	_ vestexported.VestingAccount = (*ContinuousVestingAccount)(nil)
	_ vestexported.VestingAccount = (*PeriodicVestingAccount)(nil)
	_ vestexported.VestingAccount = (*DelayedVestingAccount)(nil)
)

// Base Vesting Account

// NewBaseVestingAccount creates a new BaseVestingAccount object. It is the
// callers responsibility to ensure the base account has sufficient funds with
// regards to the original vesting amount.
func NewBaseVestingAccount(baseAccount *authtypes.BaseAccount, originalVesting sdk.Coins, endTime int64) (*BaseVestingAccount, error) {
	baseVestingAccount := &BaseVestingAccount{
		BaseAccount:      baseAccount,
		OriginalVesting:  originalVesting,
		DelegatedFree:    sdk.NewCoins(),
		DelegatedVesting: sdk.NewCoins(),
		EndTime:          endTime,
	}

	return baseVestingAccount, baseVestingAccount.Validate()
}

// LockedCoinsFromVesting returns all the coins that are not spendable (i.e. locked)
// for a vesting account given the current vesting coins. If no coins are locked,
// an empty slice of Coins is returned.
//
// CONTRACT: Delegated vesting coins and vestingCoins must be sorted.
func (bva BaseVestingAccount) LockedCoinsFromVesting(vestingCoins sdk.Coins) sdk.Coins {
	lockedCoins := vestingCoins.Sub(vestingCoins.Min(bva.DelegatedVesting)...)
	if lockedCoins == nil {
		return sdk.Coins{}
	}
	return lockedCoins
}

// TrackDelegation tracks a delegation amount for any given vesting account type
// given the amount of coins currently vesting and the current account balance
// of the delegation denominations.
//
// CONTRACT: The account's coins, delegation coins, vesting coins, and delegated
// vesting coins must be sorted.
func (bva *BaseVestingAccount) TrackDelegation(balance, vestingCoins, amount sdk.Coins) {
	for _, coin := range amount {
		baseAmt := balance.AmountOf(coin.Denom)
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
		x := math.MinInt(math.MaxInt(vestingAmt.Sub(delVestingAmt), math.ZeroInt()), coin.Amount)
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
		x := math.MinInt(delegatedFree, coin.Amount)
		y := math.MinInt(delegatedVesting, coin.Amount.Sub(x))

		if !x.IsZero() {
			xCoin := sdk.NewCoin(coin.Denom, x)
			bva.DelegatedFree = bva.DelegatedFree.Sub(xCoin)
		}

		if !y.IsZero() {
			yCoin := sdk.NewCoin(coin.Denom, y)
			bva.DelegatedVesting = bva.DelegatedVesting.Sub(yCoin)
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
	if bva.EndTime < 0 {
		return errors.New("end time cannot be negative")
	}

	if !bva.OriginalVesting.IsValid() || !bva.OriginalVesting.IsAllPositive() {
		return fmt.Errorf("invalid coins: %s", bva.OriginalVesting.String())
	}

	if !(bva.DelegatedVesting.IsAllLTE(bva.OriginalVesting)) {
		return errors.New("delegated vesting amount cannot be greater than original vesting amount")
	}

	return bva.BaseAccount.Validate()
}

// Continuous Vesting Account

var (
	_ vestexported.VestingAccount = (*ContinuousVestingAccount)(nil)
	_ authtypes.GenesisAccount    = (*ContinuousVestingAccount)(nil)
)

// NewContinuousVestingAccountRaw creates a new ContinuousVestingAccount object from BaseVestingAccount
func NewContinuousVestingAccountRaw(bva *BaseVestingAccount, startTime int64) *ContinuousVestingAccount {
	return &ContinuousVestingAccount{
		BaseVestingAccount: bva,
		StartTime:          startTime,
	}
}

// NewContinuousVestingAccount returns a new ContinuousVestingAccount
func NewContinuousVestingAccount(baseAcc *authtypes.BaseAccount, originalVesting sdk.Coins, startTime, endTime int64) (*ContinuousVestingAccount, error) {
	baseVestingAcc := &BaseVestingAccount{
		BaseAccount:     baseAcc,
		OriginalVesting: originalVesting,
		EndTime:         endTime,
	}

	continuousVestingAccount := &ContinuousVestingAccount{
		StartTime:          startTime,
		BaseVestingAccount: baseVestingAcc,
	}

	return continuousVestingAccount, continuousVestingAccount.Validate()
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
	s := math.LegacyNewDec(x).Quo(math.LegacyNewDec(y))

	for _, ovc := range cva.OriginalVesting {
		vestedAmt := math.LegacyNewDecFromInt(ovc.Amount).Mul(s).RoundInt()
		vestedCoins = append(vestedCoins, sdk.NewCoin(ovc.Denom, vestedAmt))
	}

	return vestedCoins
}

// GetVestingCoins returns the total number of vesting coins. If no coins are
// vesting, nil is returned.
func (cva ContinuousVestingAccount) GetVestingCoins(blockTime time.Time) sdk.Coins {
	return cva.OriginalVesting.Sub(cva.GetVestedCoins(blockTime)...)
}

// LockedCoins returns the set of coins that are not spendable (i.e. locked),
// defined as the vesting coins that are not delegated.
func (cva ContinuousVestingAccount) LockedCoins(ctx sdk.Context) sdk.Coins {
	return cva.BaseVestingAccount.LockedCoinsFromVesting(cva.GetVestingCoins(ctx.BlockTime()))
}

// TrackDelegation tracks a desired delegation amount by setting the appropriate
// values for the amount of delegated vesting, delegated free, and reducing the
// overall amount of base coins.
func (cva *ContinuousVestingAccount) TrackDelegation(blockTime time.Time, balance, amount sdk.Coins) {
	cva.BaseVestingAccount.TrackDelegation(balance, cva.GetVestingCoins(blockTime), amount)
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

// Periodic Vesting Account

var (
	_ vestexported.VestingAccount = (*PeriodicVestingAccount)(nil)
	_ authtypes.GenesisAccount    = (*PeriodicVestingAccount)(nil)
)

// NewPeriodicVestingAccountRaw creates a new PeriodicVestingAccount object from BaseVestingAccount
func NewPeriodicVestingAccountRaw(bva *BaseVestingAccount, startTime int64, periods Periods) *PeriodicVestingAccount {
	return &PeriodicVestingAccount{
		BaseVestingAccount: bva,
		StartTime:          startTime,
		VestingPeriods:     periods,
	}
}

// NewPeriodicVestingAccount returns a new PeriodicVestingAccount
func NewPeriodicVestingAccount(baseAcc *authtypes.BaseAccount, originalVesting sdk.Coins, startTime int64, periods Periods) (*PeriodicVestingAccount, error) {
	endTime := startTime
	for _, p := range periods {
		endTime += p.Length
	}

	baseVestingAcc := &BaseVestingAccount{
		BaseAccount:     baseAcc,
		OriginalVesting: originalVesting,
		EndTime:         endTime,
	}

	periodicVestingAccount := &PeriodicVestingAccount{
		BaseVestingAccount: baseVestingAcc,
		StartTime:          startTime,
		VestingPeriods:     periods,
	}

	return periodicVestingAccount, periodicVestingAccount.Validate()
}

// GetVestedCoins returns the total number of vested coins. If no coins are vested,
// nil is returned.
func (pva PeriodicVestingAccount) GetVestedCoins(blockTime time.Time) sdk.Coins {
	coins := ReadSchedule(pva.StartTime, pva.EndTime, pva.VestingPeriods, pva.OriginalVesting, blockTime.Unix())
	if coins.IsZero() {
		return nil
	}
	return coins
}

// GetVestingCoins returns the total number of vesting coins. If no coins are
// vesting, nil is returned.
func (pva PeriodicVestingAccount) GetVestingCoins(blockTime time.Time) sdk.Coins {
	return pva.OriginalVesting.Sub(pva.GetVestedCoins(blockTime)...)
}

// LockedCoins returns the set of coins that are not spendable (i.e. locked),
// defined as the vesting coins that are not delegated.
func (pva PeriodicVestingAccount) LockedCoins(ctx sdk.Context) sdk.Coins {
	return pva.BaseVestingAccount.LockedCoinsFromVesting(pva.GetVestingCoins(ctx.BlockTime()))
}

// TrackDelegation tracks a desired delegation amount by setting the appropriate
// values for the amount of delegated vesting, delegated free, and reducing the
// overall amount of base coins.
func (pva *PeriodicVestingAccount) TrackDelegation(blockTime time.Time, balance, amount sdk.Coins) {
	pva.BaseVestingAccount.TrackDelegation(balance, pva.GetVestingCoins(blockTime), amount)
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
	for i, p := range pva.VestingPeriods {
		if p.Length < 0 {
			return fmt.Errorf("period #%d has a negative length: %d", i, p.Length)
		}
		endTime += p.Length

		if !p.Amount.IsValid() || !p.Amount.IsAllPositive() {
			return fmt.Errorf("period #%d has invalid coins: %s", i, p.Amount.String())
		}

		originalVesting = originalVesting.Add(p.Amount...)
	}
	if endTime != pva.EndTime {
		return errors.New("vesting end time does not match length of all vesting periods")
	}
	if endTime < pva.GetStartTime() {
		return errors.New("cumulative endTime overflowed, and/or is less than startTime")
	}
	if !originalVesting.Equal(pva.OriginalVesting) {
		return fmt.Errorf("original vesting coins (%v) does not match the sum of all coins in vesting periods (%v)", pva.OriginalVesting, originalVesting)
	}

	return pva.BaseVestingAccount.Validate()
}

// Delayed Vesting Account

var (
	_ vestexported.VestingAccount = (*DelayedVestingAccount)(nil)
	_ authtypes.GenesisAccount    = (*DelayedVestingAccount)(nil)
)

// NewDelayedVestingAccountRaw creates a new DelayedVestingAccount object from BaseVestingAccount
func NewDelayedVestingAccountRaw(bva *BaseVestingAccount) *DelayedVestingAccount {
	return &DelayedVestingAccount{
		BaseVestingAccount: bva,
	}
}

// NewDelayedVestingAccount returns a DelayedVestingAccount
func NewDelayedVestingAccount(baseAcc *authtypes.BaseAccount, originalVesting sdk.Coins, endTime int64) (*DelayedVestingAccount, error) {
	baseVestingAcc := &BaseVestingAccount{
		BaseAccount:     baseAcc,
		OriginalVesting: originalVesting,
		EndTime:         endTime,
	}

	delayedVestingAccount := &DelayedVestingAccount{baseVestingAcc}

	return delayedVestingAccount, delayedVestingAccount.Validate()
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
	return dva.OriginalVesting.Sub(dva.GetVestedCoins(blockTime)...)
}

// LockedCoins returns the set of coins that are not spendable (i.e. locked),
// defined as the vesting coins that are not delegated.
func (dva DelayedVestingAccount) LockedCoins(ctx sdk.Context) sdk.Coins {
	return dva.BaseVestingAccount.LockedCoinsFromVesting(dva.GetVestingCoins(ctx.BlockTime()))
}

// TrackDelegation tracks a desired delegation amount by setting the appropriate
// values for the amount of delegated vesting, delegated free, and reducing the
// overall amount of base coins.
func (dva *DelayedVestingAccount) TrackDelegation(blockTime time.Time, balance, amount sdk.Coins) {
	dva.BaseVestingAccount.TrackDelegation(balance, dva.GetVestingCoins(blockTime), amount)
}

// GetStartTime returns zero since a delayed vesting account has no start time.
func (dva DelayedVestingAccount) GetStartTime() int64 {
	return 0
}

// Validate checks for errors on the account fields
func (dva DelayedVestingAccount) Validate() error {
	return dva.BaseVestingAccount.Validate()
}

//-----------------------------------------------------------------------------
// Permanent Locked Vesting Account

var (
	_ vestexported.VestingAccount = (*PermanentLockedAccount)(nil)
	_ authtypes.GenesisAccount    = (*PermanentLockedAccount)(nil)
)

// NewPermanentLockedAccount returns a PermanentLockedAccount
func NewPermanentLockedAccount(baseAcc *authtypes.BaseAccount, coins sdk.Coins) (*PermanentLockedAccount, error) {
	baseVestingAcc := &BaseVestingAccount{
		BaseAccount:     baseAcc,
		OriginalVesting: coins,
		EndTime:         0, // ensure EndTime is set to 0, as PermanentLockedAccount's do not have an EndTime
	}

	permanentLockedAccount := &PermanentLockedAccount{baseVestingAcc}

	return permanentLockedAccount, permanentLockedAccount.Validate()
}

// GetVestedCoins returns the total amount of vested coins for a permanent locked vesting
// account. All coins are only vested once the schedule has elapsed.
func (plva PermanentLockedAccount) GetVestedCoins(_ time.Time) sdk.Coins {
	return nil
}

// GetVestingCoins returns the total number of vesting coins for a permanent locked
// vesting account.
func (plva PermanentLockedAccount) GetVestingCoins(_ time.Time) sdk.Coins {
	return plva.OriginalVesting
}

// LockedCoins returns the set of coins that are not spendable (i.e. locked),
// defined as the vesting coins that are not delegated.
func (plva PermanentLockedAccount) LockedCoins(_ sdk.Context) sdk.Coins {
	return plva.BaseVestingAccount.LockedCoinsFromVesting(plva.OriginalVesting)
}

// TrackDelegation tracks a desired delegation amount by setting the appropriate
// values for the amount of delegated vesting, delegated free, and reducing the
// overall amount of base coins.
func (plva *PermanentLockedAccount) TrackDelegation(blockTime time.Time, balance, amount sdk.Coins) {
	plva.BaseVestingAccount.TrackDelegation(balance, plva.OriginalVesting, amount)
}

// GetStartTime returns zero since a permanent locked vesting account has no start time.
func (plva PermanentLockedAccount) GetStartTime() int64 {
	return 0
}

// GetEndTime returns a vesting account's end time, we return 0 to denote that
// a permanently locked vesting account has no end time.
func (plva PermanentLockedAccount) GetEndTime() int64 {
	return 0
}

// Validate checks for errors on the account fields
func (plva PermanentLockedAccount) Validate() error {
	if plva.EndTime > 0 {
		return errors.New("permanently vested accounts cannot have an end-time")
	}

	return plva.BaseVestingAccount.Validate()
}

func (plva PermanentLockedAccount) String() string {
	out, _ := plva.MarshalYAML()
	return out.(string)
}

type getPK interface {
	GetPubKey() cryptotypes.PubKey
}

func getPKString(g getPK) string {
	if pk := g.GetPubKey(); pk != nil {
		return pk.String()
	}
	return ""
}

func marshalYaml(i interface{}) (interface{}, error) {
	bz, err := yaml.Marshal(i)
	if err != nil {
		return nil, err
	}
	return string(bz), nil
}

// True Vesting Account

var _ vestexported.VestingAccount = (*TrueVestingAccount)(nil)
var _ authtypes.GenesisAccount = (*TrueVestingAccount)(nil)

// NewTrueVestingAccountRaw creates a new TrueVestingAccount object from BaseVestingAccount
func NewTrueVestingAccountRaw(bva *BaseVestingAccount, startTime int64, lockupPeriods, vestingPeriods Periods) *TrueVestingAccount {
	return (&TrueVestingAccount{
		BaseVestingAccount: bva,
		StartTime:          startTime,
		LockupPeriods:      lockupPeriods,
		VestingPeriods:     vestingPeriods,
	}).UpdateCombined()
}

// NewTrueVestingAccount returns a new TrueVestingAccount
func NewTrueVestingAccount(baseAcc *authtypes.BaseAccount, originalVesting sdk.Coins, startTime int64, lockupPeriods, vestingPeriods Periods) *TrueVestingAccount {
	endTime := startTime
	for _, p := range vestingPeriods {
		endTime += p.Length
	}
	baseVestingAcc := &BaseVestingAccount{
		BaseAccount:     baseAcc,
		OriginalVesting: originalVesting,
		EndTime:         endTime,
	}

	return (&TrueVestingAccount{
		BaseVestingAccount: baseVestingAcc,
		StartTime:          startTime,
		LockupPeriods:      lockupPeriods,
		VestingPeriods:     vestingPeriods,
	}).UpdateCombined()
}

func (tva *TrueVestingAccount) UpdateCombined() *TrueVestingAccount {
	start, end, combined := ConjunctPeriods(tva.StartTime, tva.StartTime, tva.LockupPeriods, tva.VestingPeriods)
	tva.StartTime = start
	tva.EndTime = end
	tva.CombinedPeriods = combined
	return tva
}

// GetVestedCoins returns the total number of vested coins. If no coins are vested,
// nil is returned.
func (tva TrueVestingAccount) GetVestedCoins(blockTime time.Time) sdk.Coins {
	// XXX consider not precomputing the combined schedule and just take the
	// min of the lockup and vesting separately. It's likely that one or the
	// other schedule will be nearly trivial, so there should be little overhead
	// in recomputing the conjunction each time.
	coins := ReadSchedule(tva.StartTime, tva.EndTime, tva.CombinedPeriods, tva.OriginalVesting, blockTime.Unix())
	if coins.IsZero() {
		return nil
	}
	return coins
}

// GetVestingCoins returns the total number of vesting coins. If no coins are
// vesting, nil is returned.
func (tva TrueVestingAccount) GetVestingCoins(blockTime time.Time) sdk.Coins {
	return tva.OriginalVesting.Sub(tva.GetVestedCoins(blockTime))
}

// LockedCoins returns the set of coins that are not spendable (i.e. locked),
// defined as the vesting coins that are not delegated.
func (tva TrueVestingAccount) LockedCoins(ctx sdk.Context) sdk.Coins {
	return tva.BaseVestingAccount.LockedCoinsFromVesting(tva.GetVestingCoins(ctx.BlockTime()))
}

// TrackDelegation tracks a desired delegation amount by setting the appropriate
// values for the amount of delegated vesting, delegated free, and reducing the
// overall amount of base coins.
func (tva *TrueVestingAccount) TrackDelegation(blockTime time.Time, balance, amount sdk.Coins) {
	tva.BaseVestingAccount.TrackDelegation(balance, tva.GetVestingCoins(blockTime), amount)
}

// GetStartTime returns the time when vesting starts for a periodic vesting
// account.
func (tva TrueVestingAccount) GetStartTime() int64 {
	return tva.StartTime
}

// GetVestingPeriods returns vesting periods associated with periodic vesting account.
func (tva TrueVestingAccount) GetVestingPeriods() Periods {
	return tva.VestingPeriods
}

// coinEq returns whether two Coins are equal.
// The IsEqual() method can panic.
func coinEq(a, b sdk.Coins) bool {
	return a.IsAllLTE(b) && b.IsAllLTE(a)
}

// Validate checks for errors on the account fields
func (tva TrueVestingAccount) Validate() error {
	if tva.GetStartTime() >= tva.GetEndTime() {
		return errors.New("vesting start-time must be before end-time")
	}

	lockupEnd := tva.StartTime
	lockupCoins := sdk.NewCoins()
	for _, p := range tva.LockupPeriods {
		lockupEnd += p.Length
		lockupCoins = lockupCoins.Add(p.Amount...)
	}
	if lockupEnd > tva.EndTime {
		return errors.New("lockup schedule extends beyond account end time")
	}
	if !coinEq(lockupCoins, tva.OriginalVesting) {
		return errors.New("original vesting coins does not match the sum of all coins in lockup periods")
	}

	vestingEnd := tva.StartTime
	vestingCoins := sdk.NewCoins()
	for _, p := range tva.VestingPeriods {
		vestingEnd += p.Length
		vestingCoins = vestingCoins.Add(p.Amount...)
	}
	if vestingEnd > tva.EndTime {
		return errors.New("vesting schedule exteds beyond account end time")
	}
	if !coinEq(vestingCoins, tva.OriginalVesting) {
		return errors.New("original vesting coins does not match the sum of all coins in vesting periods")
	}

	return tva.BaseVestingAccount.Validate()
}

func (pva TrueVestingAccount) String() string {
	out, _ := pva.MarshalYAML()
	return out.(string)
}

// MarshalYAML returns the YAML representation of a TrueVestingAccount.
func (pva TrueVestingAccount) MarshalYAML() (interface{}, error) {
	accAddr, err := sdk.AccAddressFromBech32(pva.Address)
	if err != nil {
		return nil, err
	}

	out := vestingAccountYAML{
		Address:          accAddr,
		AccountNumber:    pva.AccountNumber,
		PubKey:           getPKString(pva),
		Sequence:         pva.Sequence,
		OriginalVesting:  pva.OriginalVesting,
		DelegatedFree:    pva.DelegatedFree,
		DelegatedVesting: pva.DelegatedVesting,
		EndTime:          pva.EndTime,
		StartTime:        pva.StartTime,
		VestingPeriods:   pva.VestingPeriods,
	}
	return marshalYaml(out)
}

// ComputeClawback returns an account with all future vesting events removed,
// plus the total sum of these events. When removing the future vesting events,
// the lockup schedule will also have to be capped to keep the total sums the same.
// (But future unlocking events might be preserved if they unlock currently vested coins.)
// If the amount returned is zero, then the returned account should be unchanged.
// Does not adjust DelegatedVesting
func (tva TrueVestingAccount) ComputeClawback(clawbackTime int64) (TrueVestingAccount, sdk.Coins) {
	// Compute the truncated vesting schedule and amounts.
	// Work with the schedule as the primary data and recompute derived fields, e.g. OriginalVesting.
	t := tva.StartTime
	totalVested := sdk.NewCoins()
	totalUnvested := sdk.NewCoins()
	unvestedIdx := 0
	for i, period := range tva.VestingPeriods {
		t += period.Length
		// tie in time goes to clawback
		if t < clawbackTime {
			totalVested = totalVested.Add(period.Amount...)
			unvestedIdx = i + 1
		} else {
			totalUnvested = totalUnvested.Add(period.Amount...)
		}
	}
	newVestingPeriods := tva.VestingPeriods[:unvestedIdx]

	// To cap the unlocking schedule to the new total vested, conjunct with a limiting schedule
	capPeriods := []Period{
		{
			Length: 0,
			Amount: totalVested,
		},
	}
	_, _, newLockupPeriods := ConjunctPeriods(tva.StartTime, tva.StartTime, tva.LockupPeriods, capPeriods)

	_, _, newCombinedPeriods := ConjunctPeriods(tva.StartTime, tva.StartTime, newLockupPeriods, newVestingPeriods)

	// Now construct the new account state
	tva.OriginalVesting = totalVested
	tva.EndTime = t
	tva.LockupPeriods = newLockupPeriods
	tva.VestingPeriods = newVestingPeriods
	tva.CombinedPeriods = newCombinedPeriods
	// DelegatedVesting will be adjusted elsewhere

	return tva, totalUnvested
}

// Clawback transfers unvested tokens in a TrueVestingAccount to dest.
// Future vesting events are removed. Unstaked tokens are simply sent.
// Unbonding and staked tokens are transferred with their staking state
// intact.
func (tva TrueVestingAccount) Clawback(ctx sdk.Context, dest sdk.AccAddress, ak AccountKeeper, bk BankKeeper, sk StakingKeeper) error {
	updatedAcc, toClawBack := tva.ComputeClawback(ctx.BlockTime().Unix())
	if toClawBack.IsZero() {
		return nil
	}
	addr := updatedAcc.GetAddress()

	accPtr := &updatedAcc
	writeAcc := func() { ak.SetAccount(ctx, accPtr) }
	// Write now now so that the bank module sees unvested tokens are unlocked.
	// Note that all store writes are aborted if there is a panic, so there is
	// no danger in writing incomplete results.
	writeAcc()

	// Now that future vesting events (and associated lockup) are removed,
	// the balance of the account is unlocked and can be freely transferred.
	spendable := bk.SpendableCoins(ctx, addr)
	toXfer := coinsMin(toClawBack, spendable)
	err := bk.SendCoins(ctx, addr, dest, toXfer) // unvested tokens are be unlocked now
	if err != nil {
		return err
	}
	toClawBack = toClawBack.Sub(toXfer)
	if toClawBack.IsZero() {
		return nil
	}

	// If we need more, we'll have to transfer unbonding or bonded tokens
	// Staking is the only way unvested tokens should be missing from the bank balance.
	bondDenom := sk.BondDenom(ctx)
	// Safely subtract amt of bondDenom from coins, with a floor of zero.
	subBond := func(coins sdk.Coins, amt sdk.Int) sdk.Coins {
		coinsB := sdk.NewCoins(sdk.NewCoin(bondDenom, amt))
		return coins.Sub(coinsMin(coins, coinsB))
	}
	defer writeAcc() // write again when we're done to update DelegatedVesting

	// If we need more, transfer UnbondingDelegations.
	want := toClawBack.AmountOf(bondDenom)
	unbondings := sk.GetUnbondingDelegations(ctx, addr, math.MaxUint16)
	for _, unbonding := range unbondings {
		transferred := sk.TransferUnbonding(ctx, addr, dest, sdk.ValAddress(unbonding.ValidatorAddress), want)
		updatedAcc.DelegatedVesting = subBond(updatedAcc.DelegatedVesting, transferred)
		want = want.Sub(transferred)
		if !want.IsPositive() {
			return nil
		}
	}

	// If we need more, transfer Delegations.
	delegations := sk.GetDelegatorDelegations(ctx, addr, math.MaxUint16)
	for _, delegation := range delegations {
		validatorAddr, err := sdk.ValAddressFromBech32(delegation.ValidatorAddress)
		if err != nil {
			panic(err) // shouldn't happen
		}
		validator, found := sk.GetValidator(ctx, validatorAddr)
		if !found {
			panic("validator not found") // shoudn't happen
		}
		wantShares, err := validator.SharesFromTokensTruncated(want)
		if err != nil {
			// validator has no tokens
			continue
		}
		transferredShares := sk.TransferDelegation(ctx, addr, dest, delegation.GetValidatorAddr(), wantShares)
		// to be conservative in what we're clawing back, round transferred shares up
		transferred := validator.TokensFromSharesRoundUp(transferredShares).RoundInt()
		updatedAcc.DelegatedVesting = subBond(updatedAcc.DelegatedVesting, transferred)
		want = want.Sub(transferred)
		if !want.IsPositive() {
			// Could be slightly negative, due to rounding?
			// Don't think so, due to the precautions above.
			return nil
		}
	}

	// If we've transferred everything and still haven't transferred the desired clawback amount,
	// then the account must have most some unvested tokens from slashing.
	return nil
}

// findBalance computes the current account balance on the staking dimension.
// Returns the number of bonded, unbonding, and unbonded statking tokens.
// Rounds down when computing the bonded tokens to err on the side of vested fraction
// (smaller number of bonded tokens means vested amount covers more of them).
func (tva TrueVestingAccount) findBalance(ctx sdk.Context, bk BankKeeper, sk StakingKeeper) (bonded, unbonding, unbonded sdk.Int) {
	bondDenom := sk.BondDenom(ctx)
	unbonded = bk.GetBalance(ctx, tva.GetAddress(), bondDenom).Amount

	unbonding = sdk.ZeroInt()
	unbondings := sk.GetUnbondingDelegations(ctx, tva.GetAddress(), math.MaxUint16)
	for _, unbonding := range unbondings {
		for _, entry := range unbonding.Entries {
			unbonded = unbonded.Add(entry.Balance)
		}
	}

	bonded = sdk.ZeroInt()
	delegations := sk.GetDelegatorDelegations(ctx, tva.GetAddress(), math.MaxUint16)
	for _, delegation := range delegations {
		validatorAddr, err := sdk.ValAddressFromBech32(delegation.ValidatorAddress)
		if err != nil {
			panic(err) // shouldn't happen
		}
		validator, found := sk.GetValidator(ctx, validatorAddr)
		if !found {
			panic("validator not found") // shoudn't happen
		}
		shares := delegation.Shares
		tokens := validator.TokensFromSharesTruncated(shares).RoundInt()
		bonded = bonded.Add(tokens)
	}
	return
}

// distributeReward adds the reward to the future vesting schedule in proportion to the future vesting
// staking tokens.
func (tva TrueVestingAccount) distributeReward(ctx sdk.Context, ak AccountKeeper, bondDenom string, reward sdk.Coins) {
	now := ctx.BlockTime().Unix()
	t := tva.StartTime
	firstUnvestedPeriod := 0
	unvestedTokens := sdk.ZeroInt()
	for i, period := range tva.VestingPeriods {
		t += period.Length
		if t <= now {
			firstUnvestedPeriod = i + 1
			continue
		}
		unvestedTokens = unvestedTokens.Add(period.Amount.AmountOf(bondDenom))
	}

	runningTotReward := sdk.NewCoins()
	runningTotStaking := sdk.ZeroInt()
	for i := firstUnvestedPeriod; i < len(tva.VestingPeriods); i++ {
		period := tva.VestingPeriods[i]
		runningTotStaking = runningTotStaking.Add(period.Amount.AmountOf(bondDenom))
		runningTotRatio := runningTotStaking.ToDec().Quo(unvestedTokens.ToDec())
		targetCoins := scaleCoins(reward, runningTotRatio)
		thisReward := targetCoins.Sub(runningTotReward)
		runningTotReward = targetCoins
		period.Amount = period.Amount.Add(thisReward...)
		tva.VestingPeriods[i] = period
	}

	tva.OriginalVesting = tva.OriginalVesting.Add(reward...)
	ak.SetAccount(ctx, &tva)
}

// scaleCoins scales the given coins, rounding down.
func scaleCoins(coins sdk.Coins, scale sdk.Dec) sdk.Coins {
	scaledCoins := sdk.NewCoins()
	for _, coin := range coins {
		amt := coin.Amount.ToDec().Mul(scale).TruncateInt() // round down
		scaledCoins = scaledCoins.Add(sdk.NewCoin(coin.Denom, amt))
	}
	return scaledCoins
}

// PostReward encumbers a previously-deposited reward according to the current vesting apportionment of staking.
// Note that rewards might be unvested, but are unlocked.
func (tva TrueVestingAccount) PostReward(ctx sdk.Context, reward sdk.Coins, rak, rbk, rsk interface{}) {
	// Cast keepers to expected interfaces.
	// Necessary due to difference in expected keepers between us and caller.
	ak := rak.(AccountKeeper)
	bk := rbk.(BankKeeper)
	sk := rsk.(StakingKeeper)

	// Find the scheduled amount of vested and unvested staking tokens
	bondDenom := sk.BondDenom(ctx)
	vested := ReadSchedule(tva.StartTime, tva.EndTime, tva.VestingPeriods, tva.OriginalVesting, ctx.BlockTime().Unix()).AmountOf(bondDenom)
	unvested := tva.OriginalVesting.AmountOf(bondDenom).Sub(vested)

	if unvested.IsZero() {
		// no need to adjust the vesting schedule
		return
	}

	if vested.IsZero() {
		// all staked tokens must be unvested
		tva.distributeReward(ctx, ak, bondDenom, reward)
		return
	}

	// Find current split of account balance on staking axis
	bonded, unbonding, unbonded := tva.findBalance(ctx, bk, sk)
	total := bonded.Add(unbonding).Add(unbonded)

	// Adjust vested/unvested for the actual amount in the account (transfers, slashing)
	if unvested.GT(total) {
		// must have been reduced by slashing
		unvested = total
	}
	vested = total.Sub(unvested)

	// Now restrict to just the bonded tokens, preferring them to be vested
	if vested.GT(bonded) {
		vested = bonded
	}
	unvested = bonded.Sub(vested)

	// Compute the unvested amount of reward and add to vesting schedule
	if unvested.IsZero() {
		return
	}
	if vested.IsZero() {
		tva.distributeReward(ctx, ak, bondDenom, reward)
		return
	}
	unvestedRatio := unvested.ToDec().QuoTruncate(bonded.ToDec()) // round down
	unvestedReward := scaleCoins(reward, unvestedRatio)
	tva.distributeReward(ctx, ak, bondDenom, unvestedReward)
}
