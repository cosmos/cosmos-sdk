package auth

import (
	"errors"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
)

// Account is a standard account using a sequence number for replay protection
// and a pubkey for authentication.
type Account interface {
	GetAddress() sdk.AccAddress
	SetAddress(sdk.AccAddress) error // errors if already set.

	GetPubKey() crypto.PubKey // can return nil.
	SetPubKey(crypto.PubKey) error

	GetAccountNumber() int64
	SetAccountNumber(int64) error

	GetSequence() int64
	SetSequence(int64) error

	GetCoins() sdk.Coins
	SetCoins(sdk.Coins) error
}

// AccountDecoder unmarshals account bytes
type AccountDecoder func(accountBytes []byte) (Account, error)

var _ Account = (*BaseAccount)(nil)
var _ VestingAccount = (*ContinuousVestingAccount)(nil)
var _ VestingAccount = (*DelayTransferAccount)(nil)

//-----------------------------------------------------------
// BaseAccount

// BaseAccount - base account structure.
// Extend this by embedding this in your AppAccount.
// See the examples/basecoin/types/account.go for an example.
type BaseAccount struct {
	Address       sdk.AccAddress `json:"address"`
	Coins         sdk.Coins      `json:"coins"`
	PubKey        crypto.PubKey  `json:"public_key"`
	AccountNumber int64          `json:"account_number"`
	Sequence      int64          `json:"sequence"`
}

// Prototype function for BaseAccount
func ProtoBaseAccount() Account {
	return &BaseAccount{}
}

func NewBaseAccountWithAddress(addr sdk.AccAddress) BaseAccount {
	return BaseAccount{
		Address: addr,
	}
}

// Implements sdk.Account.
func (acc BaseAccount) GetAddress() sdk.AccAddress {
	return acc.Address
}

// Implements sdk.Account.
func (acc *BaseAccount) SetAddress(addr sdk.AccAddress) error {
	if len(acc.Address) != 0 {
		return errors.New("cannot override BaseAccount address")
	}
	acc.Address = addr
	return nil
}

// Implements sdk.Account.
func (acc BaseAccount) GetPubKey() crypto.PubKey {
	return acc.PubKey
}

// Implements sdk.Account.
func (acc *BaseAccount) SetPubKey(pubKey crypto.PubKey) error {
	acc.PubKey = pubKey
	return nil
}

// Implements sdk.Account.
func (acc *BaseAccount) GetCoins() sdk.Coins {
	return acc.Coins
}

// Implements sdk.Account.
func (acc *BaseAccount) SetCoins(coins sdk.Coins) error {
	acc.Coins = coins
	return nil
}

// Implements Account
func (acc *BaseAccount) GetAccountNumber() int64 {
	return acc.AccountNumber
}

// Implements Account
func (acc *BaseAccount) SetAccountNumber(accNumber int64) error {
	acc.AccountNumber = accNumber
	return nil
}

// Implements sdk.Account.
func (acc *BaseAccount) GetSequence() int64 {
	return acc.Sequence
}

// Implements sdk.Account.
func (acc *BaseAccount) SetSequence(seq int64) error {
	acc.Sequence = seq
	return nil
}

// VestingAccount is an account that can define a vesting schedule. Vesting coins
// can still be delegated, but only transferred after they have vested.
type VestingAccount interface {
	Account

	// Returns true if account is still vesting, else false
	//
	// CONTRACT: After account is done vesting, account behaves exactly like
	// BaseAccount.
	IsVesting(time.Time) bool

	// Calculates amount of coins that can be sent to other accounts given the
	// current blocktime.
	SendableCoins(time.Time) sdk.Coins

	// Calculates the amount of coins that are locked in the vesting account.
	LockedCoins(time.Time) sdk.Coins

	// Called on bank transfer functions (e.g. bank.SendCoins and bank.InputOutputCoins)
	// Used to track coins that are transferred in and out of vesting account
	// after initialization while account is still vesting.
	TrackTransfers(sdk.Coins)
}

// Implement Vesting Interface. Continuously vests coins linearly from
// StartTime until EndTime.
type ContinuousVestingAccount struct {
	BaseAccount
	OriginalVestingCoins sdk.Coins // coins in account on Initialization
	TransferredCoins     sdk.Coins // Net coins transferred into and out of account. May be negative.

	// StartTime and EndTime used to calculate how much of OriginalCoins is
	// unlocked at any given point.
	StartTime time.Time
	EndTime   time.Time
}

func NewContinuousVestingAccount(
	addr sdk.AccAddress, originalCoins sdk.Coins,
	startTime, endTime time.Time) ContinuousVestingAccount {

	bacc := BaseAccount{
		Address: addr,
		Coins:   originalCoins,
	}
	return ContinuousVestingAccount{
		BaseAccount:          bacc,
		OriginalVestingCoins: originalCoins,
		StartTime:            startTime,
		EndTime:              endTime,
	}
}

// Implements VestingAccount interface.
func (cva ContinuousVestingAccount) IsVesting(blockTime time.Time) bool {
	return blockTime.Unix() < cva.EndTime.Unix()
}

// Implement Vesting Account interface. Uses time in context to calculate how
// many coins has been released by vesting schedule and then accounts for
// unlocked coins that have already been transferred or delegated.
func (cva ContinuousVestingAccount) SendableCoins(blockTime time.Time) sdk.Coins {
	unlockedCoins := cva.TransferredCoins

	x := blockTime.Unix() - cva.StartTime.Unix()
	y := cva.EndTime.Unix() - cva.StartTime.Unix()
	scale := sdk.NewDec(x).Quo(sdk.NewDec(y))

	// add original coins unlocked by vesting schedule
	for _, origVestingCoin := range cva.OriginalVestingCoins {
		vAmt := sdk.NewDecFromInt(origVestingCoin.Amount).Mul(scale).RoundInt()

		// Must constrain with coins left in account since some unlocked coins may
		// have left account due to delegation.
		currentAmount := cva.GetCoins().AmountOf(origVestingCoin.Denom)

		if currentAmount.LT(vAmt) {
			vAmt = currentAmount
			// prevent double count of transferred coins
			vAmt = vAmt.Sub(cva.TransferredCoins.AmountOf(origVestingCoin.Denom))
		}

		// add non-zero coins
		if !vAmt.IsZero() {
			coin := sdk.NewCoin(origVestingCoin.Denom, vAmt)
			unlockedCoins = unlockedCoins.Plus(sdk.Coins{coin})
		}
	}

	return unlockedCoins
}

// LockedCoins returns the amount of coins that are locked in the vesting account.
func (cva ContinuousVestingAccount) LockedCoins(blockTime time.Time) sdk.Coins {
	return cva.GetCoins().Minus(cva.SendableCoins(blockTime))
}

// Implement Vesting Account. Track transfers in and out of account.
//
// CONTRACT: Send amounts must be negated.
func (cva *ContinuousVestingAccount) TrackTransfers(coins sdk.Coins) {
	cva.TransferredCoins = cva.TransferredCoins.Plus(coins)
}

// Implements Vesting Account. Vests all original coins after EndTime but keeps
// them all locked until that point.
type DelayTransferAccount struct {
	BaseAccount
	TransferredCoins sdk.Coins // Any received coins are sendable immediately

	// All coins unlocked after EndTime
	EndTime time.Time
}

func NewDelayTransferAccount(addr sdk.AccAddress, originalCoins sdk.Coins, endTime time.Time) DelayTransferAccount {
	bacc := BaseAccount{
		Address: addr,
		Coins:   originalCoins,
	}
	return DelayTransferAccount{
		BaseAccount: bacc,
		EndTime:     endTime,
	}
}

// Implements VestingAccount. It returns if the account is still vesting.
func (dta DelayTransferAccount) IsVesting(blockTime time.Time) bool {
	return blockTime.Unix() < dta.EndTime.Unix()
}

// Implements VestingAccount. If Time < EndTime return only net transferred coins
// else return all coins in account (like BaseAccount).
func (dta DelayTransferAccount) SendableCoins(blockTime time.Time) sdk.Coins {
	if blockTime.Unix() < dta.EndTime.Unix() {
		sendableCoins := dta.TransferredCoins

		// Return net transferred coins if positive, then those coins are sendable.
		for _, transCoin := range dta.TransferredCoins {
			// Must constrain with coins left in account since some unlocked coins may
			// have left account due to delegation.
			amt := sendableCoins.AmountOf(transCoin.Denom)

			currentAmount := dta.GetCoins().AmountOf(transCoin.Denom)
			if currentAmount.LT(amt) {
				delta := sdk.NewCoin(transCoin.Denom, amt.Sub(currentAmount))
				sendableCoins = sendableCoins.Minus(sdk.Coins{delta})
			}
		}

		return sendableCoins
	}

	// if EndTime has passed, DelayTransferAccount behaves like BaseAccount
	return dta.BaseAccount.GetCoins()
}

// LockedCoins returns the amount of coins that are locked in the vesting account.
func (dta DelayTransferAccount) LockedCoins(blockTime time.Time) sdk.Coins {
	return dta.GetCoins().Minus(dta.SendableCoins(blockTime))
}

// Implement Vesting Account. Track transfers in and out of account
// Send amounts must be negated
func (dta *DelayTransferAccount) TrackTransfers(coins sdk.Coins) {
	dta.TransferredCoins = dta.TransferredCoins.Plus(coins)
}
