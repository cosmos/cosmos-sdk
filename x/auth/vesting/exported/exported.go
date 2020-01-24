package exported

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
)

// VestingAccount defines an account type that vests coins via a vesting schedule.
type VestingAccount interface {
	authexported.Account

	// LockedCoins returns the set of coins that are not spendable (i.e. locked)
	// given the current account balance, only including tokens of vesting
	// denominations, and the current block time.
	LockedCoins(balance sdk.Coins, blockTime time.Time) sdk.Coins

	// Delegation and undelegation accounting that returns the resulting base
	// coins amount.
	TrackDelegation(blockTime time.Time, balance, amount sdk.Coins)
	TrackUndelegation(amount sdk.Coins)

	GetVestedCoins(blockTime time.Time) sdk.Coins
	GetVestingCoins(blockTime time.Time) sdk.Coins

	GetStartTime() int64
	GetEndTime() int64

	GetOriginalVesting() sdk.Coins
	GetDelegatedFree() sdk.Coins
	GetDelegatedVesting() sdk.Coins
}
