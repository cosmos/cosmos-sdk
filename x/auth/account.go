package auth

import (
	"time"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
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

// VestingAccount is an account that can define a vesting schedule
// Vesting coins can still be delegated, but only transferred after they have vested
type VestingAccount interface {
	Account

	// Returns true if account is still vesting, else false
	// CONTRACT: After account is done vesting, account behaves exactly like BaseAccount
    IsVesting(time.Time) bool

    // Calculates amount of coins that can be sent to other accounts given the current time
	SendableCoins(time.Time) sdk.Coins
	// Called on bank transfer functions (e.g. bank.SendCoins and bank.InputOutputCoins)
	// Used to track coins that are transferred in and out of vesting account after initialization
	TrackTransfers(sdk.Coins)
}

// Implement Vesting Interface. Continuously vests coins linearly from StartTime until EndTime
type ContinuousVestingAccount struct {
	BaseAccount
    OriginalVestingCoins sdk.Coins // Coins in account on Initialization
    TransferredCoins     sdk.Coins // Net coins transferred into and out of account. May be negative

    // StartTime and EndTime used to calculate how much of OriginalCoins is unlocked at any given point
    StartTime time.Time
    EndTime   time.Time
}

func NewContinuousVestingAccount(addr sdk.AccAddress, originalCoins sdk.Coins, startTime, endTime time.Time) ContinuousVestingAccount {
	bacc := BaseAccount{
		Address: addr,
		Coins: originalCoins,
	}
	return ContinuousVestingAccount{
		BaseAccount: bacc,
		OriginalVestingCoins: originalCoins,
		StartTime: startTime,
		EndTime: endTime,
	}
}

// Implements VestingAccount interface.
func (vacc ContinuousVestingAccount) IsVesting(blockTime time.Time) bool {
	return blockTime.Unix() < vacc.EndTime.Unix()
}

// Implement Vesting Account interface. Uses time in context to calculate how many coins
// has been released by vesting schedule and then accounts for unlocked coins that have
// already been transferred or delegated
func (vacc ContinuousVestingAccount) SendableCoins(blockTime time.Time) sdk.Coins {
	unlockedCoins := vacc.TransferredCoins
	scale := sdk.NewDec(blockTime.Unix() - vacc.StartTime.Unix()).Quo(sdk.NewDec(vacc.EndTime.Unix() - vacc.StartTime.Unix()))

	// Add original coins unlocked by vesting schedule
	for _, c := range vacc.OriginalVestingCoins {
		amt := sdk.NewDecFromInt(c.Amount).Mul(scale).RoundInt()

		// Must constrain with coins left in account
		// Since some unlocked coins may have left account due to delegation
		currentAmount := vacc.GetCoins().AmountOf(c.Denom)
		if currentAmount.LT(amt) {
			amt = currentAmount
			// prevent double count of transferred coins
			amt = amt.Sub(vacc.TransferredCoins.AmountOf(c.Denom))
		}
		
		// Add non-zero coins
		if !amt.IsZero() {
			coin := sdk.NewCoin(c.Denom, amt)
			unlockedCoins = unlockedCoins.Plus(sdk.Coins{coin})
		}
	}

	return unlockedCoins
}

// Implement Vesting Account. Track transfers in and out of account
// CONTRACT: Send amounts must be negated
func (vacc *ContinuousVestingAccount) TrackTransfers(coins sdk.Coins) {
	vacc.TransferredCoins = vacc.TransferredCoins.Plus(coins)
}

// Implements Vesting Account. Vests all original coins after EndTime but keeps them 
// all locked until that point.
type DelayTransferAccount struct {
	BaseAccount
	TransferredCoins sdk.Coins // Any received coins are sendable immediately

	// All coins unlocked after EndTime
	EndTime time.Time
}

func NewDelayTransferAccount(addr sdk.AccAddress, originalCoins sdk.Coins, endTime time.Time) DelayTransferAccount {
	bacc := BaseAccount{
		Address: addr,
		Coins: originalCoins,
	}
	return DelayTransferAccount{
		BaseAccount: bacc,
		EndTime: endTime,
	}
}

// Implements VestingAccount
func (vacc DelayTransferAccount) IsVesting(blockTime time.Time) bool {
	return blockTime.Unix() < vacc.EndTime.Unix()
}

// Implements VestingAccount. If Time < EndTime return only net transferred coins
// Else return all coins in account (like BaseAccount)
func (vacc DelayTransferAccount) SendableCoins(blockTime time.Time) sdk.Coins {
	// Check if ctx.Time < EndTime
	if blockTime.Unix() < vacc.EndTime.Unix() {
		// Return net transferred coins
		// If positive, then those coins are sendable
		sendableCoins := vacc.TransferredCoins
		for _, c := range vacc.TransferredCoins {
			// Must constrain with coins left in account
			// Since some unlocked coins may have left account due to delegation
			amt := sendableCoins.AmountOf(c.Denom)
			currentAmount := vacc.GetCoins().AmountOf(c.Denom)
			if currentAmount.LT(amt) {
				delta := sdk.Coin{c.Denom, amt.Sub(currentAmount)}
				sendableCoins = sendableCoins.Minus(sdk.Coins{delta})
			}
		}
		return sendableCoins
	}
	
	// If EndTime has passed, DelayTransferAccount behaves like BaseAccount
	return vacc.BaseAccount.GetCoins()
}

// Implement Vesting Account. Track transfers in and out of account
// Send amounts must be negated
func (vacc *DelayTransferAccount) TrackTransfers(coins sdk.Coins) {
	vacc.TransferredCoins = vacc.TransferredCoins.Plus(coins)
}

//----------------------------------------
// Wire

// Most users shouldn't use this, but this comes handy for tests.
func RegisterAccount(cdc *wire.Codec) {
	cdc.RegisterInterface((*Account)(nil), nil)
	cdc.RegisterInterface((*VestingAccount)(nil), nil)
	cdc.RegisterConcrete(&BaseAccount{}, "cosmos-sdk/BaseAccount", nil)
	cdc.RegisterConcrete(&ContinuousVestingAccount{}, "cosmos-sdk/ContinuousVestingAccount", nil)
	cdc.RegisterConcrete(&DelayTransferAccount{}, "cosmos-sdk/DelayTransferAccount", nil)
	wire.RegisterCrypto(cdc)
}
