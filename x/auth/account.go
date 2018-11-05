package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
)

// Account is an interface used to store coins at a given address within state.
// It presumes a notion of sequence numbers for replay protection,
// a notion of account numbers for replay protection for previously pruned accounts,
// and a pubkey for authentication purposes.
//
// Many complex conditions can be used in the concrete struct which implements Account.
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

// VestingAccount defines an account type that vests coins via a vesting schedule.
type VestingAccount interface {
	Account

	// Calculates the amount of coins that can be sent to other accounts given
	// the current time.
	SpendableCoins(ctx sdk.Context) sdk.Coins

	TrackDelegation(amount sdk.Coins)   // Performs delegation accounting.
	TrackUndelegation(amount sdk.Coins) // Performs undelegation accounting.
}

// AccountDecoder unmarshals account bytes
type AccountDecoder func(accountBytes []byte) (Account, error)

//-----------------------------------------------------------------------------
// BaseAccount

var _ Account = (*BaseAccount)(nil)

// BaseAccount - a base account structure.
// This can be extended by embedding within in your AppAccount.
// There are examples of this in: examples/basecoin/types/account.go.
// However one doesn't have to use BaseAccount as long as your struct
// implements Account.
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

//-----------------------------------------------------------------------------
// Vesting Accounts

// TODO: uncomment once implemented
// var (
// 	_ VestingAccount = (*ContinuousVestingAccount)(nil)
// 	_ VestingAccount = (*DelayedVestingAccount)(nil)
// )

type (
	// BaseVestingAccount implements the VestingAccount interface. It contains all
	// the necessary fields needed for any vesting account implementation.
	BaseVestingAccount struct {
		BaseAccount

		OriginalVesting sdk.Coins // coins in account upon initialization
		DelegatedFree   sdk.Coins // coins that are vested and delegated
		EndTime         time.Time // when the coins become unlocked
	}

	// ContinuousVestingAccount implements the VestingAccount interface. It
	// continuously vests by unlocking coins linearly with respect to time.
	ContinuousVestingAccount struct {
		BaseVestingAccount

		DelegatedVesting sdk.Coins // coins that vesting and delegated
		StartTime        time.Time // when the coins start to vest
	}

	// DelayedVestingAccount implements the VestingAccount interface. It vests all
	// coins after a specific time, but non prior. In other words, it keeps them
	// locked until a specified time.
	DelayedVestingAccount struct {
		BaseVestingAccount
	}
)

func NewContinuousVestingAccount(
	addr sdk.AccAddress, origCoins sdk.Coins, startTime, endTime time.Time,
) ContinuousVestingAccount {

	baseAcc := NewBaseAccountWithAddress(addr)
	baseAcc.SetCoins(origCoins)

	baseVestingAcc := BaseVestingAccount{
		BaseAccount:     baseAcc,
		OriginalVesting: origCoins,
		EndTime:         endTime,
	}

	return ContinuousVestingAccount{
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
	if blockTime.Unix() <= cva.StartTime.Unix() {
		return vestedCoins
	}

	// calculate the vesting scalar
	x := blockTime.Unix() - cva.StartTime.Unix()
	y := cva.EndTime.Unix() - cva.StartTime.Unix()
	s := sdk.NewDec(x).Quo(sdk.NewDec(y))

	for _, ovc := range cva.OriginalVesting {
		vestedAmt := sdk.NewDecFromInt(ovc.Amount).Mul(s).RoundInt()
		vestedCoin := sdk.NewCoin(ovc.Denom, vestedAmt)
		vestedCoins = vestedCoins.Plus(sdk.Coins{vestedCoin})
	}

	return vestedCoins
}

// GetVestingCoins returns the total number of vesting coins. If no coins are
// vesting, nil is returned.
func (cva ContinuousVestingAccount) GetVestingCoins(blockTime time.Time) sdk.Coins {
	return cva.OriginalVesting.Minus(cva.GetVestedCoins(blockTime))
}

// SpendableCoins returns the total number of spendable coins per denom for a
// continuous vesting account.
func (cva ContinuousVestingAccount) SpendableCoins(blockTime time.Time) sdk.Coins {
	var spendableCoins sdk.Coins

	bc := cva.GetCoins()
	v := cva.GetVestingCoins(blockTime)

	for _, coin := range bc {
		baseAmt := coin.Amount
		delVestingAmt := cva.DelegatedVesting.AmountOf(coin.Denom)
		vestingAmt := v.AmountOf(coin.Denom)

		a := baseAmt.Add(delVestingAmt)
		a = a.Sub(vestingAmt)

		var spendableCoin sdk.Coin

		// compute the min((baseAmt + delVestingAmt) - vestingAmt, baseAmt)
		fmt.Println(a)
		fmt.Println(baseAmt)
		if a.LT(baseAmt) {
			spendableCoin = sdk.NewCoin(coin.Denom, a)
		} else {
			spendableCoin = sdk.NewCoin(coin.Denom, baseAmt)
		}

		if !spendableCoin.IsZero() {
			spendableCoins = spendableCoins.Plus(sdk.Coins{spendableCoin})
		}
	}

	return spendableCoins
}

//-----------------------------------------------------------------------------
// Codec

// Most users shouldn't use this, but this comes in handy for tests.
func RegisterBaseAccount(cdc *codec.Codec) {
	cdc.RegisterInterface((*Account)(nil), nil)
	cdc.RegisterInterface((*VestingAccount)(nil), nil)
	cdc.RegisterConcrete(&BaseAccount{}, "cosmos-sdk/BaseAccount", nil)
	cdc.RegisterConcrete(&ContinuousVestingAccount{}, "cosmos-sdk/ContinuousVestingAccount", nil)
	cdc.RegisterConcrete(&DelayedVestingAccount{}, "cosmos-sdk/DelayedVestingAccount", nil)
	codec.RegisterCrypto(cdc)
}
