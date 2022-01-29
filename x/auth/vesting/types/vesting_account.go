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

// AddGrant merges a new periodic vesting grant into an existing PeriodicVestingAccount.
func (pva *PeriodicVestingAccount) AddGrant(grantStartTime int64, grantVestingPeriods []Period, grantCoins sdk.Coins) {
	newStart, newEnd, newPeriods := DisjunctPeriods(pva.StartTime, grantStartTime,
		pva.GetVestingPeriods(), grantVestingPeriods)
	pva.StartTime = newStart
	pva.EndTime = newEnd
	pva.VestingPeriods = newPeriods
	pva.OriginalVesting = pva.OriginalVesting.Add(grantCoins...)
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

// Clawback Vesting Account

var _ vestexported.VestingAccount = (*ClawbackVestingAccount)(nil)
var _ authtypes.GenesisAccount = (*ClawbackVestingAccount)(nil)

// NewClawbackVestingAccount returns a new ClawbackVestingAccount
func NewClawbackVestingAccount(baseAcc *authtypes.BaseAccount, funder sdk.AccAddress, originalVesting sdk.Coins, startTime int64, lockupPeriods, vestingPeriods Periods) *ClawbackVestingAccount {
	// copy and align schedules to avoid mutating inputs
	lp := make(Periods, len(lockupPeriods))
	copy(lp, lockupPeriods)
	vp := make(Periods, len(vestingPeriods))
	copy(vp, vestingPeriods)
	_, endTime := AlignSchedules(startTime, startTime, lp, vp)
	baseVestingAcc := &BaseVestingAccount{
		BaseAccount:     baseAcc,
		OriginalVesting: originalVesting,
		EndTime:         endTime,
	}

	return &ClawbackVestingAccount{
		BaseVestingAccount: baseVestingAcc,
		FunderAddress:      funder.String(),
		StartTime:          startTime,
		LockupPeriods:      lp,
		VestingPeriods:     vp,
	}
}

// GetVestedCoins returns the total number of vested coins. If no coins are vested,
// nil is returned.
func (va ClawbackVestingAccount) GetVestedCoins(blockTime time.Time) sdk.Coins {
	// It's likely that one or the other schedule will be nearly trivial,
	// so there should be little overhead in recomputing the conjunction each time.
	coins := coinsMin(va.GetUnlockedOnly(blockTime), va.GetVestedOnly(blockTime))
	if coins.IsZero() {
		return nil
	}
	return coins
}

// GetVestingCoins returns the total number of vesting coins. If no coins are
// vesting, nil is returned.
func (va ClawbackVestingAccount) GetVestingCoins(blockTime time.Time) sdk.Coins {
	return va.OriginalVesting.Sub(va.GetVestedCoins(blockTime))
}

// LockedCoins returns the set of coins that are not spendable (i.e. locked),
// defined as the vesting coins that are not delegated.
func (va ClawbackVestingAccount) LockedCoins(ctx sdk.Context) sdk.Coins {
	return va.BaseVestingAccount.LockedCoinsFromVesting(va.GetVestingCoins(ctx.BlockTime()))
}

// TrackDelegation tracks a desired delegation amount by setting the appropriate
// values for the amount of delegated vesting, delegated free, and reducing the
// overall amount of base coins.
func (va *ClawbackVestingAccount) TrackDelegation(blockTime time.Time, balance, amount sdk.Coins) {
	va.BaseVestingAccount.TrackDelegation(balance, va.GetVestingCoins(blockTime), amount)
}

// GetStartTime returns the time when vesting starts for a periodic vesting
// account.
func (va ClawbackVestingAccount) GetStartTime() int64 {
	return va.StartTime
}

// GetVestingPeriods returns vesting periods associated with periodic vesting account.
func (va ClawbackVestingAccount) GetVestingPeriods() Periods {
	return va.VestingPeriods
}

// coinEq returns whether two Coins are equal.
// The IsEqual() method can panic.
func coinEq(a, b sdk.Coins) bool {
	return a.IsAllLTE(b) && b.IsAllLTE(a)
}

// Validate checks for errors on the account fields
func (va ClawbackVestingAccount) Validate() error {
	if va.GetStartTime() >= va.GetEndTime() {
		return errors.New("vesting start-time must be before end-time")
	}

	lockupEnd := va.StartTime
	lockupCoins := sdk.NewCoins()
	for _, p := range va.LockupPeriods {
		lockupEnd += p.Length
		lockupCoins = lockupCoins.Add(p.Amount...)
	}
	if lockupEnd > va.EndTime {
		return errors.New("lockup schedule extends beyond account end time")
	}
	if !coinEq(lockupCoins, va.OriginalVesting) {
		return errors.New("original vesting coins does not match the sum of all coins in lockup periods")
	}

	vestingEnd := va.StartTime
	vestingCoins := sdk.NewCoins()
	for _, p := range va.VestingPeriods {
		vestingEnd += p.Length
		vestingCoins = vestingCoins.Add(p.Amount...)
	}
	if vestingEnd > va.EndTime {
		return errors.New("vesting schedule exteds beyond account end time")
	}
	if !coinEq(vestingCoins, va.OriginalVesting) {
		return errors.New("original vesting coins does not match the sum of all coins in vesting periods")
	}

	return va.BaseVestingAccount.Validate()
}

func (va ClawbackVestingAccount) String() string {
	out, _ := va.MarshalYAML()
	return out.(string)
}

// MarshalYAML returns the YAML representation of a ClawbackVestingAccount.
func (va ClawbackVestingAccount) MarshalYAML() (interface{}, error) {
	accAddr, err := sdk.AccAddressFromBech32(va.Address)
	if err != nil {
		return nil, err
	}

	out := vestingAccountYAML{
		Address:          accAddr,
		AccountNumber:    va.AccountNumber,
		PubKey:           getPKString(va),
		Sequence:         va.Sequence,
		OriginalVesting:  va.OriginalVesting,
		DelegatedFree:    va.DelegatedFree,
		DelegatedVesting: va.DelegatedVesting,
		EndTime:          va.EndTime,
		StartTime:        va.StartTime,
		VestingPeriods:   va.VestingPeriods,
	}
	return marshalYaml(out)
}

// AddGrant merges a new clawback vesting grant into an existing ClawbackVestingAccount.
func (va *ClawbackVestingAccount) AddGrant(grantStartTime int64, grantLockupPeriods, grantVestingPeriods []Period, grantCoins sdk.Coins) {
	newLockupStart, newLockupEnd, newLockupPeriods := DisjunctPeriods(va.StartTime, grantStartTime, va.LockupPeriods, grantLockupPeriods)
	newVestingStart, newVestingEnd, newVestingPeriods := DisjunctPeriods(va.StartTime, grantStartTime,
		va.GetVestingPeriods(), grantVestingPeriods)
	if newLockupStart != newVestingStart {
		panic("bad start time calculation")
	}
	va.StartTime = newLockupStart
	va.EndTime = max64(newLockupEnd, newVestingEnd)
	va.LockupPeriods = newLockupPeriods
	va.VestingPeriods = newVestingPeriods
	va.OriginalVesting = va.OriginalVesting.Add(grantCoins...)
}

// GetUnlockedOnly returns the unlocking schedule at blockTIme.
// Like GetVestedCoins, but only for the lockup component.
func (va ClawbackVestingAccount) GetUnlockedOnly(blockTime time.Time) sdk.Coins {
	return ReadSchedule(va.StartTime, va.EndTime, va.LockupPeriods, va.OriginalVesting, blockTime.Unix())
}

// GetVestedOnly returns the vesting schedule and blockTime.
// Like GetVestedCoins, but only for the vesting (in the clawback sense) component.
func (va ClawbackVestingAccount) GetVestedOnly(blockTime time.Time) sdk.Coins {
	return ReadSchedule(va.StartTime, va.EndTime, va.VestingPeriods, va.OriginalVesting, blockTime.Unix())
}

// ComputeClawback returns an account with all future vesting events removed,
// plus the total sum of these events. When removing the future vesting events,
// the lockup schedule will also have to be capped to keep the total sums the same.
// (But future unlocking events might be preserved if they unlock currently vested coins.)
// If the amount returned is zero, then the returned account should be unchanged.
// Does not adjust DelegatedVesting
func (va ClawbackVestingAccount) ComputeClawback(clawbackTime int64) (ClawbackVestingAccount, sdk.Coins) {
	// Compute the truncated vesting schedule and amounts.
	// Work with the schedule as the primary data and recompute derived fields, e.g. OriginalVesting.
	t := va.StartTime
	totalVested := sdk.NewCoins()
	totalUnvested := sdk.NewCoins()
	unvestedIdx := 0
	for i, period := range va.VestingPeriods {
		t += period.Length
		// tie in time goes to clawback
		if t < clawbackTime {
			totalVested = totalVested.Add(period.Amount...)
			unvestedIdx = i + 1
		} else {
			totalUnvested = totalUnvested.Add(period.Amount...)
		}
	}
	newVestingPeriods := va.VestingPeriods[:unvestedIdx]

	// To cap the unlocking schedule to the new total vested, conjunct with a limiting schedule
	capPeriods := []Period{
		{
			Length: 0,
			Amount: totalVested,
		},
	}
	_, _, newLockupPeriods := ConjunctPeriods(va.StartTime, va.StartTime, va.LockupPeriods, capPeriods)

	// Now construct the new account state
	va.OriginalVesting = totalVested
	va.EndTime = t
	va.LockupPeriods = newLockupPeriods
	va.VestingPeriods = newVestingPeriods
	// DelegatedVesting and DelegatedFree will be adjusted elsewhere

	return va, totalUnvested
}

// updateDelegation returns an account with its delegation bookkeeping modified for clawback,
// given the current disposition of the account's bank and staking state. Also returns
// the modified amount to claw back.
//
// Computation steps:
// - first, compute the total amount in bonded and unbonding states, used for BaseAccount bookkeeping;
// - based on the old bookkeeping, determine the amount lost to slashing since origin;
// - clip the amount to claw back to be at most the full funds in the account;
// - first claw back the unbonded funds, then go after what's delegated;
// - to the remaining delegated amount, add what's slashed;
// - the "encumbered" (locked up and/or vesting) amount of this goes in DV;
// - the remainder of the new delegated amount goes in DF.
func (va ClawbackVestingAccount) updateDelegation(encumbered, toClawBack, bonded, unbonding, unbonded sdk.Coins) (ClawbackVestingAccount, sdk.Coins) {
	delegated := bonded.Add(unbonding...)
	oldDelegated := va.DelegatedVesting.Add(va.DelegatedFree...)
	slashed := oldDelegated.Sub(coinsMin(delegated, oldDelegated))
	total := delegated.Add(unbonded...)
	toClawBack = coinsMin(toClawBack, total) // might have been slashed
	newDelegated := coinsMin(delegated, total.Sub(toClawBack)).Add(slashed...)
	va.DelegatedVesting = coinsMin(encumbered, newDelegated)
	va.DelegatedFree = newDelegated.Sub(va.DelegatedVesting)
	return va, toClawBack
}

// Clawback transfers unvested tokens in a ClawbackVestingAccount to dest.
// Future vesting events are removed. Unstaked tokens are simply sent.
// Unbonding and staked tokens are transferred with their staking state
// intact.  Account state is updated to reflect the removals.
func (va ClawbackVestingAccount) Clawback(ctx sdk.Context, dest sdk.AccAddress, ak AccountKeeper, bk BankKeeper, sk StakingKeeper) error {
	// Compute the clawback based on the account state only, and update account
	updatedAcc, toClawBack := va.ComputeClawback(ctx.BlockTime().Unix())
	if toClawBack.IsZero() {
		return nil
	}
	addr := updatedAcc.GetAddress()
	bondDenom := sk.BondDenom(ctx)

	// Compute the clawback based on bank balance and delegation, and update account
	encumbered := updatedAcc.GetVestingCoins(ctx.BlockTime())
	bondedAmt, unbondingAmt, _ := updatedAcc.findBalance(ctx, bk, sk)
	bonded := sdk.NewCoins(sdk.NewCoin(bondDenom, bondedAmt))
	unbonding := sdk.NewCoins(sdk.NewCoin(bondDenom, unbondingAmt))
	unbonded := bk.GetAllBalances(ctx, addr)
	updatedAcc, toClawBack = updatedAcc.updateDelegation(encumbered, toClawBack, bonded, unbonding, unbonded)

	// Write now now so that the bank module sees unvested tokens are unlocked.
	// Note that all store writes are aborted if there is a panic, so there is
	// no danger in writing incomplete results.
	ak.SetAccount(ctx, &updatedAcc)

	// Now that future vesting events (and associated lockup) are removed,
	// the balance of the account is unlocked and can be freely transferred.
	spendable := bk.SpendableCoins(ctx, addr)
	toXfer := coinsMin(toClawBack, spendable)
	err := bk.SendCoins(ctx, addr, dest, toXfer)
	if err != nil {
		return err // shouldn't happen, given spendable check
	}
	toClawBack = toClawBack.Sub(toXfer)

	// We need to traverse the staking data structures to update the
	// vesting account bookkeeping, and to recover more funds if necessary.
	// Staking is the only way unvested tokens should be missing from the bank balance.

	// If we need more, transfer UnbondingDelegations.
	want := toClawBack.AmountOf(bondDenom)
	unbondings := sk.GetUnbondingDelegations(ctx, addr, math.MaxUint16)
	for _, unbonding := range unbondings {
		valAddr, err := sdk.ValAddressFromBech32(unbonding.ValidatorAddress)
		if err != nil {
			panic(err)
		}
		transferred := sk.TransferUnbonding(ctx, addr, dest, valAddr, want)
		want = want.Sub(transferred)
		if !want.IsPositive() {
			break
		}
	}

	// If we need more, transfer Delegations.
	if want.IsPositive() {
		delegations := sk.GetDelegatorDelegations(ctx, addr, math.MaxUint16)
		for _, delegation := range delegations {
			validatorAddr, err := sdk.ValAddressFromBech32(delegation.ValidatorAddress)
			if err != nil {
				panic(err) // shouldn't happen
			}
			validator, found := sk.GetValidator(ctx, validatorAddr)
			if !found {
				// validator has been removed
				continue
			}
			wantShares, err := validator.SharesFromTokensTruncated(want)
			if err != nil {
				// validator has no tokens
				continue
			}
			transferredShares := sk.TransferDelegation(ctx, addr, dest, delegation.GetValidatorAddr(), wantShares)
			// to be conservative in what we're clawing back, round transferred shares up
			transferred := validator.TokensFromSharesRoundUp(transferredShares).RoundInt()
			want = want.Sub(transferred)
			if !want.IsPositive() {
				// Could be slightly negative, due to rounding?
				// Don't think so, due to the precautions above.
				break
			}
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
func (va ClawbackVestingAccount) findBalance(ctx sdk.Context, bk BankKeeper, sk StakingKeeper) (bonded, unbonding, unbonded sdk.Int) {
	bondDenom := sk.BondDenom(ctx)
	unbonded = bk.GetBalance(ctx, va.GetAddress(), bondDenom).Amount

	unbonding = sdk.ZeroInt()
	unbondings := sk.GetUnbondingDelegations(ctx, va.GetAddress(), math.MaxUint16)
	for _, ubd := range unbondings {
		for _, entry := range ubd.Entries {
			unbonding = unbonding.Add(entry.Balance)
		}
	}

	bonded = sdk.ZeroInt()
	delegations := sk.GetDelegatorDelegations(ctx, va.GetAddress(), math.MaxUint16)
	for _, delegation := range delegations {
		validatorAddr, err := sdk.ValAddressFromBech32(delegation.ValidatorAddress)
		if err != nil {
			panic(err) // shouldn't happen
		}
		validator, found := sk.GetValidator(ctx, validatorAddr)
		if !found {
			// validator has been removed
			continue
		}
		shares := delegation.Shares
		tokens := validator.TokensFromSharesTruncated(shares).RoundInt()
		bonded = bonded.Add(tokens)
	}
	return bonded, unbonding, unbonded
}

// distributeReward adds the reward to the future vesting schedule in proportion to the future vesting
// staking tokens.
func (va ClawbackVestingAccount) distributeReward(ctx sdk.Context, ak AccountKeeper, bondDenom string, reward sdk.Coins) {
	now := ctx.BlockTime().Unix()
	t := va.StartTime
	firstUnvestedPeriod := 0
	unvestedTokens := sdk.ZeroInt()
	for i, period := range va.VestingPeriods {
		t += period.Length
		if t <= now {
			firstUnvestedPeriod = i + 1
			continue
		}
		unvestedTokens = unvestedTokens.Add(period.Amount.AmountOf(bondDenom))
	}

	runningTotReward := sdk.NewCoins()
	runningTotStaking := sdk.ZeroInt()
	for i := firstUnvestedPeriod; i < len(va.VestingPeriods); i++ {
		period := va.VestingPeriods[i]
		runningTotStaking = runningTotStaking.Add(period.Amount.AmountOf(bondDenom))
		runningTotRatio := runningTotStaking.ToDec().Quo(unvestedTokens.ToDec())
		targetCoins := scaleCoins(reward, runningTotRatio)
		thisReward := targetCoins.Sub(runningTotReward)
		runningTotReward = targetCoins
		period.Amount = period.Amount.Add(thisReward...)
		va.VestingPeriods[i] = period
	}

	va.OriginalVesting = va.OriginalVesting.Add(reward...)
	ak.SetAccount(ctx, &va)
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

// minInt returns the minumum of its arguments.
func minInt(a, b sdk.Int) sdk.Int {
	if a.GT(b) {
		return b
	}
	return a
}

// PostReward encumbers a previously-deposited reward according to the current vesting apportionment of staking.
// Note that rewards might be unvested, but are unlocked.
func (va ClawbackVestingAccount) PostReward(ctx sdk.Context, reward sdk.Coins, ak AccountKeeper, bk BankKeeper, sk StakingKeeper) {
	// Find the scheduled amount of vested and unvested staking tokens
	bondDenom := sk.BondDenom(ctx)
	vested := ReadSchedule(va.StartTime, va.EndTime, va.VestingPeriods, va.OriginalVesting, ctx.BlockTime().Unix()).AmountOf(bondDenom)
	unvested := va.OriginalVesting.AmountOf(bondDenom).Sub(vested)

	if unvested.IsZero() {
		// no need to adjust the vesting schedule
		return
	}

	if vested.IsZero() {
		// all staked tokens must be unvested
		va.distributeReward(ctx, ak, bondDenom, reward)
		return
	}

	// Find current split of account balance on staking axis
	bonded, unbonding, unbonded := va.findBalance(ctx, bk, sk)
	total := bonded.Add(unbonding).Add(unbonded)
	total = total.Sub(minInt(total, reward.AmountOf(bondDenom))) // look at pre-reward total

	// Adjust vested/unvested for the actual amount in the account (transfers, slashing)
	// preferring them to be unvested
	unvested = minInt(unvested, total) // may have been reduced by slashing
	vested = total.Sub(unvested)

	// Now restrict to just the bonded tokens, preferring them to be vested
	vested = minInt(vested, bonded)
	unvested = bonded.Sub(vested)

	// Compute the unvested amount of reward and add to vesting schedule
	if unvested.IsZero() {
		return
	}
	if vested.IsZero() {
		va.distributeReward(ctx, ak, bondDenom, reward)
		return
	}
	unvestedRatio := unvested.ToDec().QuoTruncate(bonded.ToDec()) // round down
	unvestedReward := scaleCoins(reward, unvestedRatio)
	va.distributeReward(ctx, ak, bondDenom, unvestedReward)
}
